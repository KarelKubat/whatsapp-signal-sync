package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

type SyncEngine struct {
	cfg                   *Config
	waClient              *WhatsAppClient
	sig                   *SignalClient
	done                  chan struct{}
	personalSignalGroupID string

	stateFilePath string
	state         *State
	stateMu       sync.Mutex

	// Deduplication caches to prevent echo loops in linked groups
	waSentIDs     map[string]time.Time
	sigSentTimes  map[int64]time.Time
	cacheMu       sync.Mutex
}

func NewSyncEngine(cfg *Config, waClient *WhatsAppClient, sigClient *SignalClient) *SyncEngine {
	return &SyncEngine{
		cfg:          cfg,
		waClient:     waClient,
		sig:          sigClient,
		done:         make(chan struct{}),
		waSentIDs:    make(map[string]time.Time),
		sigSentTimes: make(map[int64]time.Time),
	}
}

func (s *SyncEngine) Start(ctx context.Context) {
	// Ensure temp attachment directory exists
	if s.cfg.Storage.TempAttachmentDir != "" {
		_ = os.MkdirAll(s.cfg.Storage.TempAttachmentDir, 0700)
	}

	// Try to find the Signal group named "Whatsapp Signal Sync"
	sigGroups, err := s.sig.ListGroups(ctx)
	if err == nil {
		for _, g := range sigGroups {
			if strings.ToLower(g.Name) == "whatsapp signal sync" {
				s.personalSignalGroupID = g.ID
				log.Printf("[SyncEngine] Found Signal group 'Whatsapp Signal Sync' (ID: %s). Direct personal forwards will be routed here.", g.ID)
				break
			}
		}
	}
	if s.personalSignalGroupID == "" {
		log.Println("[SyncEngine] No 'Whatsapp Signal Sync' Signal group found. Defaulting personal forwards to 'Note to Self' (your Signal number).")
	}

	// Load State
	statePath := filepath.Join(filepath.Dir(s.cfg.Storage.WhatsAppDB), "state.json")
	s.stateFilePath = statePath
	var errState error
	s.state, errState = LoadState(statePath)
	if errState != nil {
		log.Printf("[SyncEngine] Failed to load state: %v. Missed messages sync might be skipped.", errState)
	} else {
		log.Printf("[SyncEngine] Loaded state: WhatsApp Last Timestamp: %d, Signal Last Timestamp: %d", s.state.LastWhatsAppTimestamp, s.state.LastSignalTimestamp)
	}

	go s.listenWhatsApp(ctx)
	go s.listenSignal(ctx)
	go s.periodicSyncLoop(ctx)
	go s.cleanCacheLoop(ctx)
}

func (s *SyncEngine) Stop() {
	close(s.done)
}

func (s *SyncEngine) addWASentID(id string) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.waSentIDs[id] = time.Now()
}

func (s *SyncEngine) isWASent(id string) bool {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	_, exists := s.waSentIDs[id]
	if exists {
		delete(s.waSentIDs, id)
		return true
	}
	return false
}

func (s *SyncEngine) addSigSentTime(t int64) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.sigSentTimes[t] = time.Now()
}

func (s *SyncEngine) isSigSent(t int64) bool {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	_, exists := s.sigSentTimes[t]
	if exists {
		delete(s.sigSentTimes, t)
		return true
	}
	return false
}

func (s *SyncEngine) cleanCacheLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cacheMu.Lock()
			now := time.Now()
			for id, t := range s.waSentIDs {
				if now.Sub(t) > 1*time.Minute {
					delete(s.waSentIDs, id)
				}
			}
			for timeKey, t := range s.sigSentTimes {
				if now.Sub(t) > 1*time.Minute {
					delete(s.sigSentTimes, timeKey)
				}
			}
			s.cacheMu.Unlock()
		}
	}
}

func (s *SyncEngine) periodicSyncLoop(ctx context.Context) {
	// Trigger an initial receive on startup to pull any missed messages immediately
	log.Println("[SyncEngine] Triggering initial Signal receive catch-up...")
	if err := s.sig.TriggerReceive(ctx); err != nil {
		if isAlreadyReceivingError(err) {
			log.Println("[SyncEngine] Signal daemon is already actively receiving messages (connection is healthy).")
		} else {
			log.Printf("[SyncEngine] Initial Signal receive trigger failed: %v", err)
		}
	}

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("[SyncEngine] Running periodic Signal receive catch-up...")
			if err := s.sig.TriggerReceive(ctx); err != nil {
				if isAlreadyReceivingError(err) {
					log.Println("[SyncEngine] Signal daemon is already actively receiving messages (connection is healthy).")
				} else {
					log.Printf("[SyncEngine] Periodic Signal receive trigger failed: %v", err)
				}
			}
		}
	}
}

func (s *SyncEngine) listenWhatsApp(ctx context.Context) {
	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case msg := <-s.waClient.incomingEvents:
			if s.cfg.Debug {
				payload, _ := json.MarshalIndent(msg, "", "  ")
				log.Printf("[DEBUG] WhatsApp Event Received on Channel:\n%s", string(payload))
				if msg.Message != nil {
					log.Printf("[DEBUG] Raw msg.Message struct: %+v", msg.Message)
				} else {
					log.Printf("[DEBUG] msg.Message is nil")
				}
			}
			// Discard messages sent by the sync engine itself to prevent loops
			if s.isWASent(msg.Info.ID) {
				log.Printf("[Sync WhatsApp -> Signal] Discarding message sent by self: %s", msg.Info.ID)
				continue
			}

			// Discard status messages / broadcast messages
			if msg.Info.Chat.Server == "broadcast" {
				continue
			}

			s.handleWhatsAppMessage(ctx, msg)
		}
	}
}

func (s *SyncEngine) handleWhatsAppMessage(ctx context.Context, msg *events.Message) {
	msgTime := msg.Info.Timestamp.Unix()
	s.stateMu.Lock()
	if s.state != nil {
		isDuplicate := false
		if msgTime < s.state.LastWhatsAppTimestamp {
			isDuplicate = true
		} else if msgTime == s.state.LastWhatsAppTimestamp && msg.Info.ID == s.state.LastWhatsAppMsgID {
			isDuplicate = true
		}
		if isDuplicate {
			s.stateMu.Unlock()
			log.Printf("[Sync WhatsApp -> Signal] Discarding duplicate or older message (msgTime: %d, lastTime: %d, msgID: %s, lastID: %s)", msgTime, s.state.LastWhatsAppTimestamp, msg.Info.ID, s.state.LastWhatsAppMsgID)
			return
		}
	}
	s.stateMu.Unlock()

	log.Printf("[Sync WhatsApp -> Signal] Received message JID: %s, Sender: %s, ID: %s", msg.Info.Chat.String(), msg.Info.Sender.String(), msg.Info.ID)

	// Extract message text and media
	var text string
	var imageMsg *waE2E.ImageMessage
	var videoMsg *waE2E.VideoMessage

	if msg.Message.Conversation != nil {
		text = msg.Message.GetConversation()
	} else if msg.Message.ExtendedTextMessage != nil {
		text = msg.Message.GetExtendedTextMessage().GetText()
	} else if msg.Message.ImageMessage != nil {
		imageMsg = msg.Message.ImageMessage
		text = imageMsg.GetCaption()
	} else if msg.Message.VideoMessage != nil {
		videoMsg = msg.Message.VideoMessage
		text = videoMsg.GetCaption()
	} else {
		// Fallback for unsupported messages (like Polls)
		text = "[WhatsApp received a message type that cannot be forwarded. Check your personal WhatsApp.]"
	}

	// Prepare attachments
	var attachments []string
	if imageMsg != nil {
		filePath, err := s.downloadWhatsAppMedia(ctx, imageMsg, "img_"+uuid.New().String()+".jpg")
		if err != nil {
			log.Printf("Failed to download WhatsApp image: %v", err)
			text = fmt.Sprintf("[WhatsApp image forward failed: %v]", err)
		} else {
			attachments = append(attachments, filePath)
			defer os.Remove(filePath)
		}
	} else if videoMsg != nil {
		filePath, err := s.downloadWhatsAppMedia(ctx, videoMsg, "vid_"+uuid.New().String()+".mp4")
		if err != nil {
			log.Printf("Failed to download WhatsApp video: %v", err)
			text = fmt.Sprintf("[WhatsApp video forward failed: %v]", err)
		} else {
			attachments = append(attachments, filePath)
			defer os.Remove(filePath)
		}
	}

	senderName := msg.Info.PushName
	if senderName == "" {
		senderName = msg.Info.Sender.User
	}

	// Forwarding targets
	var signalRecipient string
	var signalGroup string
	var formattedText string

	// Check if this WhatsApp message is a reply (quote) to a forwarded Signal message
	quotedText := getQuotedMessageText(msg)
	isReply := false
	if quotedText != "" {
		if sigNum := parsePrefix(quotedText, "[Signal Direct: ", "]"); sigNum != "" {
			signalRecipient = sigNum
			formattedText = text // send raw reply
			isReply = true
			log.Printf("[Sync WhatsApp -> Signal] Detected reply to Signal Direct number: %s", sigNum)
		} else if sigGroupID := parsePrefix(quotedText, "[Signal Group: ", "]"); sigGroupID != "" {
			// Check if this Signal group is linked!
			isLinked := false
			for _, targetSigID := range s.cfg.GroupLinks {
				if targetSigID == sigGroupID {
					isLinked = true
					break
				}
			}

			if isLinked {
				signalGroup = sigGroupID
				formattedText = text // send raw reply
				isReply = true
				log.Printf("[Sync WhatsApp -> Signal] Detected reply to linked Signal Group ID: %s", sigGroupID)
			} else {
				// Unlinked! Fail the reply and route back to personal account with warning header
				if s.personalSignalGroupID != "" {
					signalGroup = s.personalSignalGroupID
				} else {
					signalRecipient = s.cfg.Accounts.SignalNumber
				}
				formattedText = fmt.Sprintf("[Signal Group Reply Failed] Could not reply to unlinked Signal group %s. Your reply: %s", sigGroupID, text)
				isReply = true
				log.Printf("[Sync WhatsApp -> Signal] Blocked reply to unlinked Signal Group ID: %s", sigGroupID)
			}
		}
	}

	if !isReply {
		if msg.Info.IsFromMe {
			isLinkedGroup := false
			if msg.Info.IsGroup {
				_, isLinkedGroup = s.cfg.GroupLinks[msg.Info.Chat.String()]
			}
			if !isLinkedGroup {
				log.Printf("[Sync WhatsApp -> Signal] Discarding non-reply message from self in non-linked chat: Chat JID: %s, ID: %s", msg.Info.Chat.String(), msg.Info.ID)
				return
			}
		}

		if msg.Info.IsGroup {
			// Group message forwarding
			waGroupID := msg.Info.Chat.String()
			sigGroupID, linked := s.cfg.GroupLinks[waGroupID]

			if linked {
				signalGroup = sigGroupID
				if s.cfg.Debug {
					formattedText = formatForwardText(fmt.Sprintf("[WhatsApp Group: %s]", waGroupID), senderName, text)
				} else {
					formattedText = formatForwardText("", senderName, text)
				}
			} else {
				// Unlinked group, forward to personal account
				formattedText = formatForwardText(fmt.Sprintf("[WhatsApp Group: %s]", waGroupID), senderName, text)
				if s.personalSignalGroupID != "" {
					signalGroup = s.personalSignalGroupID
				} else {
					signalRecipient = s.cfg.Accounts.SignalNumber
				}
			}
		} else {
			// Personal message forwarding
			formattedText = formatForwardText(fmt.Sprintf("[WhatsApp Direct: %s]", msg.Info.Sender.String()), senderName, text)
			if s.personalSignalGroupID != "" {
				signalGroup = s.personalSignalGroupID
			} else {
				signalRecipient = s.cfg.Accounts.SignalNumber
			}
		}
	}

	// Send to Signal
	if signalRecipient != "" || signalGroup != "" {
		if s.cfg.Debug {
			log.Printf("[DEBUG] Forwarding to Signal: recipient=%s, group=%s, message=%q, attachments=%v", signalRecipient, signalGroup, formattedText, attachments)
		}
		sentTime, err := s.sig.SendMessage(ctx, signalRecipient, signalGroup, formattedText, attachments)
		if err != nil {
			log.Printf("Failed to forward WhatsApp message to Signal: %v", err)
		} else {
			log.Printf("Forwarded WhatsApp message to Signal successfully.")
			if sentTime > 0 {
				s.addSigSentTime(sentTime)
			}

			// Update state
			s.stateMu.Lock()
			if s.state != nil {
				s.state.LastWhatsAppTimestamp = msgTime
				s.state.LastWhatsAppMsgID = msg.Info.ID
				if err := SaveState(s.stateFilePath, s.state); err != nil {
					log.Printf("[SyncEngine] Failed to save state: %v", err)
				}
			}
			s.stateMu.Unlock()
		}
	}
}

func (s *SyncEngine) downloadWhatsAppMedia(ctx context.Context, mediaMsg whatsmeow.DownloadableMessage, filename string) (string, error) {
	data, err := s.waClient.DownloadMedia(ctx, mediaMsg)
	if err != nil {
		return "", err
	}

	tmpPath := filepath.Join(s.cfg.Storage.TempAttachmentDir, filename)
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return "", err
	}
	return tmpPath, nil
}

func (s *SyncEngine) listenSignal(ctx context.Context) {
	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case event := <-s.sig.incomingEvents:
			if s.cfg.Debug {
				payload, _ := json.MarshalIndent(event, "", "  ")
				log.Printf("[DEBUG] Signal Event Received on Socket:\n%s", string(payload))
			}
			// Discard messages sent by the sync engine itself to prevent loops
			if s.isSigSent(event.Params.Envelope.Timestamp) {
				log.Printf("[Sync Signal -> WhatsApp] Discarding message sent by self: %d", event.Params.Envelope.Timestamp)
				continue
			}

			msgContent := event.Params.Envelope.DataMessage
			if msgContent == nil && event.Params.Envelope.SyncMessage != nil {
				msgContent = event.Params.Envelope.SyncMessage.SentMessage
			}

			if msgContent == nil {
				continue
			}

			s.handleSignalMessage(ctx, event)
		}
	}
}

func (s *SyncEngine) handleSignalMessage(ctx context.Context, event *SignalMessageEvent) {
	msgTimeMs := event.Params.Envelope.Timestamp
	s.stateMu.Lock()
	if s.state != nil && msgTimeMs <= s.state.LastSignalTimestamp {
		s.stateMu.Unlock()
		log.Printf("[Sync Signal -> WhatsApp] Discarding message older than last sync timestamp (msgMs: %d, lastMs: %d)", msgTimeMs, s.state.LastSignalTimestamp)
		return
	}
	s.stateMu.Unlock()

	msg := event.Params.Envelope.DataMessage
	if msg == nil && event.Params.Envelope.SyncMessage != nil {
		msg = event.Params.Envelope.SyncMessage.SentMessage
	}
	log.Printf("[Sync Signal -> WhatsApp] Received message from Source: %s, Msg: %s", event.Params.Envelope.SourceNumber, msg.Message)

	// Discard empty text messages without attachments
	if msg.Message == "" && len(msg.Attachments) == 0 {
		return
	}

	senderName := event.Params.Envelope.SourceName
	if senderName == "" {
		senderName = event.Params.Envelope.SourceNumber
	}
	if senderName == "" {
		senderName = event.Params.Envelope.SourceUUID
	}

	// Prepare JID target
	var whatsappTarget JID
	var formattedText string
	isReply := false

	// Check if this Signal message is a reply (quote) to a forwarded WhatsApp message
	if msg.Quote != nil && msg.Quote.Text != "" {
		quotedText := msg.Quote.Text
		if waJIDStr := parsePrefix(quotedText, "[WhatsApp Direct: ", "]"); waJIDStr != "" {
			targetJID, err := ParseWhatsAppJID(waJIDStr)
			if err == nil {
				whatsappTarget = targetJID
				formattedText = msg.Message // send raw reply
				isReply = true
				log.Printf("[Sync Signal -> WhatsApp] Detected reply to WhatsApp Direct JID: %s", waJIDStr)
			}
		} else if waGroupIDStr := parsePrefix(quotedText, "[WhatsApp Group: ", "]"); waGroupIDStr != "" {
			targetJID, err := ParseWhatsAppJID(waGroupIDStr)
			if err == nil {
				// Check if this WhatsApp group is linked!
				_, linked := s.cfg.GroupLinks[targetJID.String()]
				if linked {
					whatsappTarget = targetJID
					formattedText = msg.Message // send raw reply
					isReply = true
					log.Printf("[Sync Signal -> WhatsApp] Detected reply to linked WhatsApp Group JID: %s", waGroupIDStr)
				} else {
					// Unlinked! Fail the reply and route back to personal account with warning header
					whatsappTarget = JID{Raw: s.cfg.Accounts.WhatsAppUserJID, IsGroup: false}
					formattedText = fmt.Sprintf("[WhatsApp Group Reply Failed] Could not reply to unlinked WhatsApp group %s. Your reply: %s", waGroupIDStr, msg.Message)
					isReply = true
					log.Printf("[Sync Signal -> WhatsApp] Blocked reply to unlinked WhatsApp Group JID: %s", waGroupIDStr)
				}
			}
		}
	}

	if !isReply {
		if event.Params.Envelope.SourceNumber == s.cfg.Accounts.SignalNumber {
			isLinkedGroup := false
			if msg.GroupInfo != nil && msg.GroupInfo.GroupID != "" {
				sigGroupID := msg.GroupInfo.GroupID
				for _, targetSigID := range s.cfg.GroupLinks {
					if targetSigID == sigGroupID {
						isLinkedGroup = true
						break
					}
				}
			}
			if !isLinkedGroup {
				log.Printf("[Sync Signal -> WhatsApp] Discarding non-reply message from self in non-linked chat: Source: %s", event.Params.Envelope.SourceNumber)
				return
			}
		}

		if msg.GroupInfo != nil && msg.GroupInfo.GroupID != "" {
			// Group message forwarding
			sigGroupID := msg.GroupInfo.GroupID
			var linkedWAJID string
			var isLinked bool

			for waJID, targetSigID := range s.cfg.GroupLinks {
				if targetSigID == sigGroupID {
					linkedWAJID = waJID
					isLinked = true
					break
				}
			}

			if isLinked {
				whatsappTarget = JID{Raw: linkedWAJID, IsGroup: true}
				if s.cfg.Debug {
					formattedText = formatForwardText(fmt.Sprintf("[Signal Group: %s]", sigGroupID), senderName, msg.Message)
				} else {
					formattedText = formatForwardText("", senderName, msg.Message)
				}
			} else {
				// Unlinked group, forward to personal contact
				whatsappTarget = JID{Raw: s.cfg.Accounts.WhatsAppUserJID, IsGroup: false}
				formattedText = formatForwardText(fmt.Sprintf("[Signal Group: %s] %s (in %s)", sigGroupID, senderName, msg.GroupInfo.Name), "", msg.Message)
			}
		} else {
			// Personal message forwarding
			whatsappTarget = JID{Raw: s.cfg.Accounts.WhatsAppUserJID, IsGroup: false}
			formattedText = formatForwardText(fmt.Sprintf("[Signal Direct: %s]", event.Params.Envelope.SourceNumber), senderName, msg.Message)
		}
	}

	// Check for attachments (images/video)
	hasAttachmentsFailed := false
	var sentID string
	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			// Check if it is an image or video
			isVideo := strings.HasPrefix(attachment.ContentType, "video/")
			isImage := strings.HasPrefix(attachment.ContentType, "image/")

			if isImage || isVideo {
				filePath := attachment.StoredFilename
				if filePath == "" && attachment.ID != "" {
					filePath = filepath.Join(s.cfg.Storage.SignalConfigDir, "attachments", attachment.ID)
					if s.cfg.Debug {
						log.Printf("[DEBUG] StoredFilename was empty. Trying fallback path: %s", filePath)
					}
				}

				data, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Failed to read Signal attachment (path: %s): %v", filePath, err)
					_, _ = s.waClient.SendTextMessage(ctx, whatsappTarget, fmt.Sprintf("[Signal attachment forward failed: %v]", err))
					hasAttachmentsFailed = true
					continue
				}

				if s.cfg.Debug {
					log.Printf("[DEBUG] Forwarding media to WhatsApp: target=%+v, MIME=%s, isVideo=%t, text=%q, dataLength=%d", whatsappTarget, attachment.ContentType, isVideo, formattedText, len(data))
				}
				sentID, err = s.waClient.SendMediaMessage(ctx, whatsappTarget, data, attachment.ContentType, isVideo, formattedText)
				if err != nil {
					log.Printf("Failed to forward Signal media message: %v", err)
					hasAttachmentsFailed = true
				} else {
					if sentID != "" {
						s.addWASentID(sentID)
					}
				}
			} else {
				// Unsupported attachment type, notify user
				_, _ = s.waClient.SendTextMessage(ctx, whatsappTarget, fmt.Sprintf("%s\n[Signal received an unsupported attachment type (%s). Check your personal Signal.]", formattedText, attachment.ContentType))
			}
		}
	} else {
		// Simple text message forwarding
		var err error
		if s.cfg.Debug {
			log.Printf("[DEBUG] Forwarding text to WhatsApp: target=%+v, message=%q", whatsappTarget, formattedText)
		}
		sentID, err = s.waClient.SendTextMessage(ctx, whatsappTarget, formattedText)
		if err != nil {
			log.Printf("Failed to forward Signal message to WhatsApp: %v", err)
			hasAttachmentsFailed = true
		} else {
			log.Printf("Forwarded Signal message to WhatsApp successfully.")
			if sentID != "" {
				s.addWASentID(sentID)
			}
		}
	}

	// Update state only if forwarding succeeded
	if !hasAttachmentsFailed {
		s.stateMu.Lock()
		if s.state != nil {
			s.state.LastSignalTimestamp = msgTimeMs
			if err := SaveState(s.stateFilePath, s.state); err != nil {
				log.Printf("[SyncEngine] Failed to save state: %v", err)
			}
		}
		s.stateMu.Unlock()
	}
}

// Helpers for parsing prefixes and JIDs in replies
func parsePrefix(text, patternStart, patternEnd string) string {
	startIdx := strings.Index(text, patternStart)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(patternStart)
	endIdx := strings.Index(text[startIdx:], patternEnd)
	if endIdx == -1 {
		return ""
	}
	return text[startIdx : startIdx+endIdx]
}

func getQuotedMessageText(msg *events.Message) string {
	if msg.Message == nil || msg.Message.ExtendedTextMessage == nil || msg.Message.ExtendedTextMessage.ContextInfo == nil {
		return ""
	}
	ctxInfo := msg.Message.ExtendedTextMessage.ContextInfo
	if ctxInfo.QuotedMessage == nil {
		return ""
	}
	qm := ctxInfo.QuotedMessage
	if qm.Conversation != nil {
		return qm.GetConversation()
	}
	if qm.ExtendedTextMessage != nil {
		return qm.GetExtendedTextMessage().GetText()
	}
	if qm.ImageMessage != nil {
		return qm.ImageMessage.GetCaption()
	}
	if qm.VideoMessage != nil {
		return qm.VideoMessage.GetCaption()
	}
	return ""
}

func isAlreadyReceivingError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "already being received")
}

func formatForwardText(headerPrefix, senderName, messageText string) string {
	var prefix string
	if headerPrefix != "" {
		prefix = headerPrefix + " "
	}
	if messageText == "" {
		return fmt.Sprintf("%s%s", prefix, senderName)
	}
	return fmt.Sprintf("%s%s: %s", prefix, senderName, messageText)
}
