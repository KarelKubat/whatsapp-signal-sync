package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
}

func NewSyncEngine(cfg *Config, waClient *WhatsAppClient, sigClient *SignalClient) *SyncEngine {
	return &SyncEngine{
		cfg:      cfg,
		waClient: waClient,
		sig:      sigClient,
		done:     make(chan struct{}),
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

	go s.listenWhatsApp(ctx)
	go s.listenSignal(ctx)
}

func (s *SyncEngine) Stop() {
	close(s.done)
}

func (s *SyncEngine) listenWhatsApp(ctx context.Context) {
	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case msg := <-s.waClient.incomingEvents:
			// Discard messages sent by self to avoid infinite loops
			if msg.Info.IsFromMe {
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
	log.Printf("[Sync WhatsApp -> Signal] Received message JID: %s, Sender: %s", msg.Info.Chat.String(), msg.Info.Sender.String())

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
			signalGroup = sigGroupID
			formattedText = text // send raw reply
			isReply = true
			log.Printf("[Sync WhatsApp -> Signal] Detected reply to Signal Group ID: %s", sigGroupID)
		}
	}

	if !isReply {
		if msg.Info.IsGroup {
			// Group message forwarding
			waGroupID := msg.Info.Chat.String()
			sigGroupID, linked := s.cfg.GroupLinks[waGroupID]

			if linked {
				signalGroup = sigGroupID
				formattedText = fmt.Sprintf("%s: %s", senderName, text)
			} else {
				// Unlinked group, forward to personal account
				formattedText = fmt.Sprintf("[WhatsApp Group: %s] %s: %s", waGroupID, senderName, text)
				if s.personalSignalGroupID != "" {
					signalGroup = s.personalSignalGroupID
				} else {
					signalRecipient = s.cfg.Accounts.SignalNumber
				}
			}
		} else {
			// Personal message forwarding
			formattedText = fmt.Sprintf("[WhatsApp Direct: %s] %s: %s", msg.Info.Sender.String(), senderName, text)
			if s.personalSignalGroupID != "" {
				signalGroup = s.personalSignalGroupID
			} else {
				signalRecipient = s.cfg.Accounts.SignalNumber
			}
		}
	}

	// Send to Signal
	if signalRecipient != "" || signalGroup != "" {
		err := s.sig.SendMessage(ctx, signalRecipient, signalGroup, formattedText, attachments)
		if err != nil {
			log.Printf("Failed to forward WhatsApp message to Signal: %v", err)
		} else {
			log.Printf("Forwarded WhatsApp message to Signal successfully.")
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
			// Discard messages sent by self
			if event.Envelope.SourceNumber == s.cfg.Accounts.SignalNumber {
				continue
			}
			if event.Envelope.DataMessage == nil {
				continue
			}

			s.handleSignalMessage(ctx, event)
		}
	}
}

func (s *SyncEngine) handleSignalMessage(ctx context.Context, event *SignalMessageEvent) {
	msg := event.Envelope.DataMessage
	log.Printf("[Sync Signal -> WhatsApp] Received message from Source: %s, Msg: %s", event.Envelope.SourceNumber, msg.Message)

	// Discard empty text messages without attachments
	if msg.Message == "" && len(msg.Attachments) == 0 {
		return
	}

	senderName := event.Envelope.SourceName
	if senderName == "" {
		senderName = event.Envelope.SourceNumber
	}
	if senderName == "" {
		senderName = event.Envelope.SourceUUID
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
				whatsappTarget = targetJID
				formattedText = msg.Message // send raw reply
				isReply = true
				log.Printf("[Sync Signal -> WhatsApp] Detected reply to WhatsApp Group JID: %s", waGroupIDStr)
			}
		}
	}

	if !isReply {
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
				formattedText = fmt.Sprintf("%s: %s", senderName, msg.Message)
			} else {
				// Unlinked group, forward to personal contact
				whatsappTarget = JID{Raw: s.cfg.Accounts.WhatsAppUserJID, IsGroup: false}
				formattedText = fmt.Sprintf("[Signal Group: %s] %s (in %s): %s", sigGroupID, senderName, msg.GroupInfo.Name, msg.Message)
			}
		} else {
			// Personal message forwarding
			whatsappTarget = JID{Raw: s.cfg.Accounts.WhatsAppUserJID, IsGroup: false}
			formattedText = fmt.Sprintf("[Signal Direct: %s] %s: %s", event.Envelope.SourceNumber, senderName, msg.Message)
		}
	}

	// Check for attachments (images/video)
	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			// Check if it is an image or video
			isVideo := strings.HasPrefix(attachment.ContentType, "video/")
			isImage := strings.HasPrefix(attachment.ContentType, "image/")

			if isImage || isVideo {
				data, err := os.ReadFile(attachment.StoredFilename)
				if err != nil {
					log.Printf("Failed to read Signal attachment: %v", err)
					_ = s.waClient.SendTextMessage(ctx, whatsappTarget, fmt.Sprintf("[Signal attachment forward failed: %v]", err))
					continue
				}

				err = s.waClient.SendMediaMessage(ctx, whatsappTarget, data, attachment.ContentType, isVideo, formattedText)
				if err != nil {
					log.Printf("Failed to forward Signal media message: %v", err)
				}
			} else {
				// Unsupported attachment type, notify user
				_ = s.waClient.SendTextMessage(ctx, whatsappTarget, fmt.Sprintf("%s\n[Signal received an unsupported attachment type (%s). Check your personal Signal.]", formattedText, attachment.ContentType))
			}
		}
	} else {
		// Simple text message forwarding
		err := s.waClient.SendTextMessage(ctx, whatsappTarget, formattedText)
		if err != nil {
			log.Printf("Failed to forward Signal message to WhatsApp: %v", err)
		} else {
			log.Printf("Forwarded Signal message to WhatsApp successfully.")
		}
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
