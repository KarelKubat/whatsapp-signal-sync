package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"database/sql"
	"encoding/base64"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
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

	// Cache maps for group names
	waGroupNames  map[string]string
	sigGroupNames map[string]string
	namesMu       sync.Mutex
}

func NewSyncEngine(cfg *Config, waClient *WhatsAppClient, sigClient *SignalClient) *SyncEngine {
	return &SyncEngine{
		cfg:           cfg,
		waClient:      waClient,
		sig:           sigClient,
		done:          make(chan struct{}),
		waSentIDs:     make(map[string]time.Time),
		sigSentTimes:  make(map[int64]time.Time),
		waGroupNames:  make(map[string]string),
		sigGroupNames: make(map[string]string),
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
		s.namesMu.Lock()
		for _, g := range sigGroups {
			s.sigGroupNames[g.ID] = g.Name
			if strings.ToLower(g.Name) == "whatsapp signal sync" && s.personalSignalGroupID == "" {
				s.personalSignalGroupID = g.ID
				log.Printf("[SyncEngine] Found Signal group 'Whatsapp Signal Sync' (ID: %s). Direct personal forwards will be routed here.", g.ID)
			}
		}
		s.namesMu.Unlock()
	}
	if s.personalSignalGroupID == "" {
		log.Println("[SyncEngine] No 'Whatsapp Signal Sync' Signal group found. Defaulting personal forwards to 'Note to Self' (your Signal number).")
	}

	// Populate WhatsApp groups cache
	waGroups, err := s.waClient.GetGroups(ctx)
	if err == nil {
		s.namesMu.Lock()
		for _, g := range waGroups {
			s.waGroupNames[g.JID.String()] = g.Name
		}
		s.namesMu.Unlock()
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
		if _, processed := s.state.WhatsAppProcessedIDs[msg.Info.ID]; processed {
			isDuplicate = true
		} else if msgTime < s.state.LastWhatsAppTimestamp-3600 {
			isDuplicate = true
		}
		if isDuplicate {
			s.stateMu.Unlock()
			log.Printf("[Sync WhatsApp -> Signal] Discarding duplicate or older message (msgTime: %d, lastTime: %d, msgID: %s)", msgTime, s.state.LastWhatsAppTimestamp, msg.Info.ID)
			return
		}
	}
	s.stateMu.Unlock()

	// Discard messages from archived WhatsApp chats
	if s.isWhatsAppChatArchived(ctx, msg.Info.Chat.String()) {
		if s.cfg.Debug {
			log.Printf("[DEBUG] Discarding message from archived WhatsApp chat JID: %s", msg.Info.Chat.String())
		}
		return
	}

	if msg.Message != nil && msg.Message.ProtocolMessage != nil {
		if s.cfg.Debug {
			log.Printf("[DEBUG] WhatsApp message contains ProtocolMessage (type: %s), ignoring.", msg.Message.ProtocolMessage.GetType())
		}
		return
	}

	log.Printf("[Sync WhatsApp -> Signal] Received message JID: %s, Sender: %s, ID: %s", msg.Info.Chat.String(), msg.Info.Sender.String(), msg.Info.ID)

	// Extract message text and media
	var text string
	var imageMsg *waE2E.ImageMessage
	var videoMsg *waE2E.VideoMessage
	var audioMsg *waE2E.AudioMessage

	if s.cfg.Debug {
		log.Printf("[DEBUG] handleWhatsAppMessage: msg.Message=%+v", msg.Message)
		log.Printf("[DEBUG] Condition checks: Conversation=%t (GetConversation=%q), ExtendedTextMessage=%t (GetText=%q), AudioMessage=%t, SecretEncryptedMessage=%t",
			msg.Message.Conversation != nil,
			msg.Message.GetConversation(),
			msg.Message.ExtendedTextMessage != nil,
			msg.Message.GetExtendedTextMessage().GetText(),
			msg.Message.AudioMessage != nil,
			msg.Message.SecretEncryptedMessage != nil,
		)
	}

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
	} else if msg.Message.AudioMessage != nil {
		audioMsg = msg.Message.AudioMessage
		text = ""
	} else if msg.Message.SecretEncryptedMessage != nil {
		encType := msg.Message.SecretEncryptedMessage.GetSecretEncType()
		if encType == waE2E.SecretEncryptedMessage_MESSAGE_EDIT {
			text = "[System: This message was edited on WhatsApp. Check there.]"
		} else {
			text = fmt.Sprintf("[System: WhatsApp received an encrypted message notification (type: %s).]", encType.String())
		}
	} else if msg.Message.PollCreationMessage != nil ||
		msg.Message.PollCreationMessageV2 != nil ||
		msg.Message.PollCreationMessageV3 != nil ||
		msg.Message.PollCreationMessageV4 != nil ||
		msg.Message.PollCreationMessageV5 != nil ||
		msg.Message.PollCreationMessageV6 != nil ||
		msg.Message.LocationMessage != nil ||
		msg.Message.LiveLocationMessage != nil ||
		msg.Message.ContactMessage != nil ||
		msg.Message.ContactsArrayMessage != nil ||
		msg.Message.TemplateMessage != nil ||
		msg.Message.InteractiveMessage != nil ||
		msg.Message.ButtonsMessage != nil {
		// Fallback for unsupported user-facing messages (like Polls)
		text = "[WhatsApp received a message type that cannot be forwarded. Check there.]"
	} else {
		// Silently ignore internal protocol/system messages (like SenderKeyDistributionMessage, ReactionMessage, etc.)
		if s.cfg.Debug {
			log.Printf("[DEBUG] Silently ignoring WhatsApp protocol/system message (ID: %s)", msg.Info.ID)
		}
		return
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
	} else if audioMsg != nil {
		ext := ".ogg"
		if audioMsg.GetMimetype() == "audio/mp4" || audioMsg.GetMimetype() == "audio/aac" {
			ext = ".m4a"
		} else if audioMsg.GetMimetype() == "audio/mpeg" {
			ext = ".mp3"
		}
		filePath, err := s.downloadWhatsAppMedia(ctx, audioMsg, "aud_"+uuid.New().String()+ext)
		if err != nil {
			log.Printf("Failed to download WhatsApp audio: %v", err)
			text = fmt.Sprintf("[WhatsApp audio forward failed: %v]", err)
		} else {
			attachments = append(attachments, filePath)
			defer os.Remove(filePath)
		}
	}

	senderName := s.waClient.ResolveJIDName(ctx, msg.Info.Sender)
	isNameKnown := senderName != msg.Info.Sender.User
	if !isNameKnown {
		if msg.Info.PushName != "" {
			senderName = msg.Info.PushName
			isNameKnown = true
		}
	}

	// Forwarding targets
	var signalRecipient string
	var signalGroup string
	var formattedText string
	var msgPrefix string
	var cleanText string

	// Check if this WhatsApp message is a reply (quote)
	quotedAuthorJID, quotedTextVal, quotedStanzaID := getQuotedInfo(msg)
	isReply := false
	isRoutedReply := false
	if quotedTextVal != "" {
		isReply = true
		s.stateMu.Lock()
		mapping, found := s.state.WhatsAppReplies[quotedStanzaID]
		s.stateMu.Unlock()
		if found {
			signalRecipient = mapping.SignalRecipient
			signalGroup = mapping.SignalGroup
			formattedText = text // send raw reply
			isRoutedReply = true
			log.Printf("[Sync WhatsApp -> Signal] Detected reply via database lookup. SignalRecipient: %s, SignalGroup: %s", signalRecipient, signalGroup)
		} else if isSelfJID(quotedAuthorJID, s.cfg.Accounts.WhatsAppUserJID) {
			if sigNum := parsePrefix(quotedTextVal, "[Signal Direct: ", "]"); sigNum != "" {
				signalRecipient = sigNum
				formattedText = text // send raw reply
				isRoutedReply = true
				log.Printf("[Sync WhatsApp -> Signal] Detected reply to Signal Direct number: %s", sigNum)
			} else if sigGroupID := parsePrefix(quotedTextVal, "[Signal Group: ", "]"); sigGroupID != "" {
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
					isRoutedReply = true
					log.Printf("[Sync WhatsApp -> Signal] Detected reply to linked Signal Group ID: %s", sigGroupID)
				} else {
					// Unlinked! Fail the reply and route back to personal account with warning header
					if s.personalSignalGroupID != "" {
						signalGroup = s.personalSignalGroupID
					} else {
						signalRecipient = s.cfg.Accounts.SignalNumber
					}
					formattedText = fmt.Sprintf("[Signal Group Reply Failed] Could not reply to unlinked Signal group %s. Your reply: %s", sigGroupID, text)
					isRoutedReply = true
					log.Printf("[Sync WhatsApp -> Signal] Blocked reply to unlinked Signal Group ID: %s", sigGroupID)
				}
			}
		}
	}

	if isRoutedReply {
		cleanText = formattedText
	}

	if !isRoutedReply {
		if msg.Info.IsFromMe {
			isLinkedGroup := false
			if msg.Info.IsGroup {
				_, isLinkedGroup = s.cfg.GroupLinks[msg.Info.Chat.String()]
			}
			if !isLinkedGroup && !isReply {
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
				cleanText = formatForwardText("", senderName, text)
			} else {
				// Unlinked group, forward to personal account
				waGroupName := s.getWhatsAppGroupName(ctx, waGroupID)
				msgPrefix = fmt.Sprintf("[%s]", waGroupName)
				cleanText = formatForwardText("", senderName, text)
				if s.personalSignalGroupID != "" {
					signalGroup = s.personalSignalGroupID
				} else {
					signalRecipient = s.cfg.Accounts.SignalNumber
				}
			}
		} else {
			// Personal message forwarding
			if !isNameKnown {
				msgPrefix = fmt.Sprintf("[WhatsApp Direct: %s]", msg.Info.Sender.String())
			}
			cleanText = formatForwardText("", senderName, text)
			if s.personalSignalGroupID != "" {
				signalGroup = s.personalSignalGroupID
			} else {
				signalRecipient = s.cfg.Accounts.SignalNumber
			}
		}
	}

	// Append blockquote context for replies/quotes if present
	if quotedTextVal != "" {
		quoteBlock := s.formatQuote(ctx, quotedAuthorJID, quotedTextVal)
		if msgPrefix != "" {
			formattedText = msgPrefix + "\n" + quoteBlock + "\n" + cleanText
		} else {
			formattedText = quoteBlock + "\n" + cleanText
		}
	} else {
		if msgPrefix != "" {
			formattedText = msgPrefix + " " + cleanText
		} else {
			formattedText = cleanText
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
				if msgTime > s.state.LastWhatsAppTimestamp {
					s.state.LastWhatsAppTimestamp = msgTime
				}
				s.state.WhatsAppProcessedIDs[msg.Info.ID] = time.Now().Unix()
				s.cleanUpWhatsAppProcessedIDs()

				if sentTime > 0 {
					sigTsKey := fmt.Sprintf("%d", sentTime)
					s.state.SignalReplies[sigTsKey] = MessageMapping{
						WhatsAppMsgID:   msg.Info.ID,
						WhatsAppChatJID: msg.Info.Chat.String(),
						Timestamp:       time.Now().Unix(),
					}
					s.state.CleanUpReplies()
				}

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

	// Resolve the target Signal Chat JID to check archive status
	var chatJID JID
	if msg.GroupInfo != nil && msg.GroupInfo.GroupID != "" {
		chatJID = JID{Raw: msg.GroupInfo.GroupID, IsGroup: true}
	} else {
		source := event.Params.Envelope.SourceNumber
		if event.Params.Envelope.SyncMessage != nil && event.Params.Envelope.SyncMessage.SentMessage != nil {
			sm := event.Params.Envelope.SyncMessage.SentMessage
			source = sm.DestinationNumber
			if source == "" {
				source = sm.DestinationUuid
			}
			if source == "" {
				source = sm.Destination
			}
		}
		if source == "" {
			source = event.Params.Envelope.SourceUUID
		}
		chatJID = JID{Raw: source, IsGroup: false}
	}

	if s.isSignalChatArchived(ctx, chatJID) {
		if s.cfg.Debug {
			log.Printf("[DEBUG] Discarding message from archived Signal chat JID: %s", chatJID.Raw)
		}
		return
	}

	log.Printf("[Sync Signal -> WhatsApp] Received message from Source: %s, Msg: %s", event.Params.Envelope.SourceNumber, msg.Message)

	// Discard empty text messages without attachments
	if msg.Message == "" && len(msg.Attachments) == 0 {
		return
	}

	senderName := event.Params.Envelope.SourceName
	isNameKnown := senderName != "" && senderName != event.Params.Envelope.SourceNumber && senderName != event.Params.Envelope.SourceUUID
	if senderName == "" {
		senderName = event.Params.Envelope.SourceNumber
	}
	if senderName == "" {
		senderName = event.Params.Envelope.SourceUUID
	}

	// Prepare JID target
	var whatsappTarget JID
	var formattedText string
	var msgPrefix string
	var cleanText string
	isReply := false

	// Check if this Signal message is a reply (quote) to a forwarded WhatsApp message
	isRoutedReply := false
	if msg.Quote != nil && msg.Quote.Text != "" {
		isReply = true
		quotedTimestamp := msg.Quote.ID
		sigTsKey := fmt.Sprintf("%d", quotedTimestamp)
		s.stateMu.Lock()
		mapping, found := s.state.SignalReplies[sigTsKey]
		s.stateMu.Unlock()
		if found {
			isGroup := strings.HasSuffix(mapping.WhatsAppChatJID, "@g.us")
			whatsappTarget = JID{Raw: mapping.WhatsAppChatJID, IsGroup: isGroup}
			formattedText = msg.Message // send raw reply
			isRoutedReply = true
			log.Printf("[Sync Signal -> WhatsApp] Detected reply via database lookup. WhatsAppTarget: %s", mapping.WhatsAppChatJID)
		} else if msg.Quote.Author == s.cfg.Accounts.SignalNumber {
			quotedText := msg.Quote.Text
			if waJIDStr := parsePrefix(quotedText, "[WhatsApp Direct: ", "]"); waJIDStr != "" {
				targetJID, err := ParseWhatsAppJID(waJIDStr)
				if err == nil {
					whatsappTarget = targetJID
					formattedText = msg.Message // send raw reply
					isRoutedReply = true
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
						isRoutedReply = true
						log.Printf("[Sync Signal -> WhatsApp] Detected reply to linked WhatsApp Group JID: %s", waGroupIDStr)
					} else {
						// Unlinked! Fail the reply and route back to personal account with warning header
						whatsappTarget = JID{Raw: s.cfg.Accounts.WhatsAppUserJID, IsGroup: false}
						formattedText = fmt.Sprintf("[WhatsApp Group Reply Failed] Could not reply to unlinked WhatsApp group %s. Your reply: %s", waGroupIDStr, msg.Message)
						isRoutedReply = true
						log.Printf("[Sync Signal -> WhatsApp] Blocked reply to unlinked WhatsApp Group JID: %s", waGroupIDStr)
					}
				}
			}
		}
	}

	if isRoutedReply {
		cleanText = formattedText
	}

	if !isRoutedReply {
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
			if !isLinkedGroup && !isReply {
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
				cleanText = formatForwardText("", senderName, msg.Message)
			} else {
				// Unlinked group, forward to personal contact
				whatsappTarget = JID{Raw: s.cfg.Accounts.WhatsAppUserJID, IsGroup: false}
				sigGroupName := msg.GroupInfo.Name
				if sigGroupName == "" {
					sigGroupName = s.getSignalGroupName(ctx, sigGroupID)
				}
				msgPrefix = fmt.Sprintf("[%s]", sigGroupName)
				cleanText = formatForwardText("", senderName, msg.Message)
			}
		} else {
			// Personal message forwarding
			whatsappTarget = JID{Raw: s.cfg.Accounts.WhatsAppUserJID, IsGroup: false}
			if !isNameKnown {
				msgPrefix = fmt.Sprintf("[Signal Direct: %s]", event.Params.Envelope.SourceNumber)
			}
			cleanText = formatForwardText("", senderName, msg.Message)
		}
	}

	// Append blockquote context for replies/quotes if present
	if msg.Quote != nil && msg.Quote.Text != "" {
		quoteBlock := s.formatQuote(ctx, msg.Quote.Author, msg.Quote.Text)
		if msgPrefix != "" {
			formattedText = msgPrefix + "\n" + quoteBlock + "\n" + cleanText
		} else {
			formattedText = quoteBlock + "\n" + cleanText
		}
	} else {
		if msgPrefix != "" {
			formattedText = msgPrefix + " " + cleanText
		} else {
			formattedText = cleanText
		}
	}

	// Check for attachments (images/video)
	hasAttachmentsFailed := false
	var sentID string
	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			// Check if it is an image or video or audio
			isVideo := strings.HasPrefix(attachment.ContentType, "video/")
			isImage := strings.HasPrefix(attachment.ContentType, "image/")
			isAudio := strings.HasPrefix(attachment.ContentType, "audio/")

			if isImage || isVideo || isAudio {
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

				if isAudio {
					if s.cfg.Debug {
						log.Printf("[DEBUG] Forwarding audio to WhatsApp: target=%+v, MIME=%s, dataLength=%d", whatsappTarget, attachment.ContentType, len(data))
					}
					sentID, err = s.waClient.SendAudioMessage(ctx, whatsappTarget, data, attachment.ContentType)
				} else {
					if s.cfg.Debug {
						log.Printf("[DEBUG] Forwarding media to WhatsApp: target=%+v, MIME=%s, isVideo=%t, text=%q, dataLength=%d", whatsappTarget, attachment.ContentType, isVideo, formattedText, len(data))
					}
					sentID, err = s.waClient.SendMediaMessage(ctx, whatsappTarget, data, attachment.ContentType, isVideo, formattedText)
				}

				if err != nil {
					log.Printf("Failed to forward Signal media/audio message: %v", err)
					hasAttachmentsFailed = true
				} else {
					if sentID != "" {
						s.addWASentID(sentID)
					}
				}
			} else {
				// Unsupported attachment type, notify user
				_, _ = s.waClient.SendTextMessage(ctx, whatsappTarget, fmt.Sprintf("%s\n[Signal received an unsupported attachment type (%s). Check there.]", formattedText, attachment.ContentType))
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

			if sentID != "" {
				var sigGroupID string
				if msg.GroupInfo != nil {
					sigGroupID = msg.GroupInfo.GroupID
				}
				s.state.WhatsAppReplies[sentID] = MessageMapping{
					SignalRecipient: event.Params.Envelope.SourceNumber,
					SignalGroup:     sigGroupID,
					Timestamp:       time.Now().Unix(),
				}
				s.state.CleanUpReplies()
			}

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

func getQuotedInfo(msg *events.Message) (authorJID string, quotedText string, stanzaID string) {
	if msg.Message == nil {
		return "", "", ""
	}
	var ctxInfo *waE2E.ContextInfo
	if msg.Message.ExtendedTextMessage != nil {
		ctxInfo = msg.Message.ExtendedTextMessage.ContextInfo
	} else if msg.Message.ImageMessage != nil {
		ctxInfo = msg.Message.ImageMessage.ContextInfo
	} else if msg.Message.VideoMessage != nil {
		ctxInfo = msg.Message.VideoMessage.ContextInfo
	} else if msg.Message.AudioMessage != nil {
		ctxInfo = msg.Message.AudioMessage.ContextInfo
	}

	if ctxInfo == nil || ctxInfo.QuotedMessage == nil {
		return "", "", ""
	}

	authorJID = ctxInfo.GetParticipant()
	stanzaID = ctxInfo.GetStanzaID()
	qm := ctxInfo.QuotedMessage
	if qm.Conversation != nil {
		quotedText = qm.GetConversation()
	} else if qm.ExtendedTextMessage != nil {
		quotedText = qm.GetExtendedTextMessage().GetText()
	} else if qm.ImageMessage != nil {
		quotedText = qm.ImageMessage.GetCaption()
	} else if qm.VideoMessage != nil {
		quotedText = qm.VideoMessage.GetCaption()
	}
	return authorJID, quotedText, stanzaID
}

func (s *SyncEngine) formatQuote(ctx context.Context, author, text string) string {
	if text == "" {
		return ""
	}
	if strings.HasPrefix(text, "[") {
		return formatBlockquote("", text)
	}

	cleanAuthor := author
	if parts := strings.Split(author, "@"); len(parts) > 0 {
		cleanAuthor = parts[0]
	}

	cleanSelfJID := s.cfg.Accounts.WhatsAppUserJID
	if parts := strings.Split(s.cfg.Accounts.WhatsAppUserJID, "@"); len(parts) > 0 {
		cleanSelfJID = parts[0]
	}

	if cleanAuthor == cleanSelfJID || cleanAuthor == s.cfg.Accounts.SignalNumber {
		// If it's sent by self (or the sync engine), check if the text already starts with a prefix like "Sender: "
		hasSenderPrefix := false
		colonIdx := strings.Index(text, ": ")
		if colonIdx > 0 && colonIdx < 30 {
			namePart := text[:colonIdx]
			if !strings.Contains(namePart, "\n") {
				hasSenderPrefix = true
			}
		}
		if hasSenderPrefix {
			return formatBlockquote("", text)
		}
		return formatBlockquote("Me", text)
	}

	// Resolve WhatsApp JID to a name if possible
	resolvedAuthor := cleanAuthor
	if strings.Contains(author, "@") || strings.Contains(author, ":") { // Valid JID format
		jid, err := types.ParseJID(author)
		if err == nil {
			resolvedAuthor = s.waClient.ResolveJIDName(ctx, jid)
		}
	} else if len(author) > 5 { // Might be a phone number without JID suffix
		// Try parsing as user JID
		jid, err := types.ParseJID(author + "@s.whatsapp.net")
		if err == nil {
			name := s.waClient.ResolveJIDName(ctx, jid)
			if name != jid.User {
				resolvedAuthor = name
			}
		}
	}

	return formatBlockquote(resolvedAuthor, text)
}

func formatBlockquote(author, text string) string {
	lines := strings.Split(text, "\n")
	var quotedLines []string
	if author != "" {
		quotedLines = append(quotedLines, "> "+author+": "+lines[0])
		for _, line := range lines[1:] {
			quotedLines = append(quotedLines, "> "+line)
		}
	} else {
		for _, line := range lines {
			quotedLines = append(quotedLines, "> "+line)
		}
	}
	return strings.Join(quotedLines, "\n")
}

func (s *SyncEngine) cleanUpWhatsAppProcessedIDs() {
	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	for id, ts := range s.state.WhatsAppProcessedIDs {
		if ts < cutoff {
			delete(s.state.WhatsAppProcessedIDs, id)
		}
	}
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

func isSelfJID(jid string, selfJID string) bool {
	cleanJID := jid
	if parts := strings.Split(jid, "@"); len(parts) > 0 {
		cleanJID = parts[0]
	}
	cleanSelf := selfJID
	if parts := strings.Split(selfJID, "@"); len(parts) > 0 {
		cleanSelf = parts[0]
	}
	return cleanJID == cleanSelf
}

func (s *SyncEngine) getWhatsAppGroupName(ctx context.Context, waGroupID string) string {
	s.namesMu.Lock()
	name, exists := s.waGroupNames[waGroupID]
	s.namesMu.Unlock()
	if exists {
		return name
	}

	jid, err := types.ParseJID(waGroupID)
	if err == nil {
		info, err := s.waClient.GetGroupInfo(ctx, jid)
		if err == nil && info != nil {
			s.namesMu.Lock()
			s.waGroupNames[waGroupID] = info.Name
			s.namesMu.Unlock()
			return info.Name
		}
	}
	return waGroupID
}

func (s *SyncEngine) getSignalGroupName(ctx context.Context, sigGroupID string) string {
	s.namesMu.Lock()
	name, exists := s.sigGroupNames[sigGroupID]
	s.namesMu.Unlock()
	if exists {
		return name
	}

	groups, err := s.sig.ListGroups(ctx)
	if err == nil {
		s.namesMu.Lock()
		for _, g := range groups {
			s.sigGroupNames[g.ID] = g.Name
			if g.ID == sigGroupID {
				name = g.Name
			}
		}
		s.namesMu.Unlock()
	}
	if name != "" {
		return name
	}
	return sigGroupID
}

func (s *SyncEngine) getSignalDBPath() (string, error) {
	accountsJsonPath := filepath.Join(s.cfg.Storage.SignalConfigDir, "data", "accounts.json")
	data, err := os.ReadFile(accountsJsonPath)
	if err != nil {
		return "", err
	}

	var schema struct {
		Accounts []struct {
			Number      string `json:"number"`
			StorageName string `json:"storageName"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		return "", err
	}

	var storageName string
	for _, acc := range schema.Accounts {
		if acc.Number == s.cfg.Accounts.SignalNumber {
			storageName = acc.StorageName
			break
		}
	}

	if storageName == "" {
		return "", fmt.Errorf("signal account %s not found in accounts.json", s.cfg.Accounts.SignalNumber)
	}

	return filepath.Join(s.cfg.Storage.SignalConfigDir, "data", storageName+".d", "account.db"), nil
}

func (s *SyncEngine) isSignalChatArchived(ctx context.Context, target JID) bool {
	dbPath, err := s.getSignalDBPath()
	if err != nil {
		log.Printf("[SyncEngine] Failed to resolve Signal DB path: %v", err)
		return false
	}

	// Open read-only WAL connection
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro&_journal_mode=WAL", dbPath))
	if err != nil {
		log.Printf("[SyncEngine] Failed to open Signal DB: %v", err)
		return false
	}
	defer db.Close()

	var archived int
	if target.IsGroup {
		groupIdBytes, err := base64.StdEncoding.DecodeString(target.Raw)
		if err != nil {
			return false
		}
		query := `
			SELECT r.archived 
			FROM recipient r 
			JOIN group_v2 g ON r.storage_id = g.storage_id 
			WHERE g.group_id = ?;
		`
		err = db.QueryRowContext(ctx, query, groupIdBytes).Scan(&archived)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false
			}
			log.Printf("[SyncEngine] Failed to query Signal group archive status: %v", err)
			return false
		}
	} else {
		query := `SELECT archived FROM recipient WHERE number = ? OR aci = ?;`
		err = db.QueryRowContext(ctx, query, target.Raw, target.Raw).Scan(&archived)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false
			}
			log.Printf("[SyncEngine] Failed to query Signal recipient archive status: %v", err)
			return false
		}
	}

	return archived == 1
}

func (s *SyncEngine) isWhatsAppChatArchived(ctx context.Context, waGroupID string) bool {
	jid, err := types.ParseJID(waGroupID)
	if err != nil {
		return false
	}
	return s.waClient.IsChatArchived(ctx, jid)
}
