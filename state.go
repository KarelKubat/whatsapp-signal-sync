package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type MessageMapping struct {
	WhatsAppMsgID   string `json:"whatsapp_msg_id,omitempty"`
	WhatsAppChatJID string `json:"whatsapp_chat_jid,omitempty"`
	SignalRecipient string `json:"signal_recipient,omitempty"`
	SignalGroup     string `json:"signal_group,omitempty"`
	Timestamp       int64  `json:"timestamp"`
}

type State struct {
	LastWhatsAppTimestamp int64                     `json:"last_whatsapp_timestamp"`
	LastWhatsAppMsgID     string                    `json:"last_whatsapp_msg_id"`
	LastSignalTimestamp   int64                     `json:"last_signal_timestamp"`
	WhatsAppProcessedIDs  map[string]int64          `json:"whatsapp_processed_ids"`
	WhatsAppReplies       map[string]MessageMapping `json:"whatsapp_replies"` // key: WhatsApp message ID (StanzaID)
	SignalReplies         map[string]MessageMapping `json:"signal_replies"`   // key: Signal message timestamp (string)
}

func LoadState(path string) (*State, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		// First run: default to current time to avoid importing ancient history
		now := time.Now().Unix()
		state := &State{
			LastWhatsAppTimestamp: now,
			LastSignalTimestamp:   now * 1000,
			WhatsAppProcessedIDs:  make(map[string]int64),
			WhatsAppReplies:       make(map[string]MessageMapping),
			SignalReplies:         make(map[string]MessageMapping),
		}
		if errDir := os.MkdirAll(filepath.Dir(path), 0700); errDir != nil {
			return nil, errDir
		}
		return state, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	var state State
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil {
		return nil, err
	}

	if state.WhatsAppProcessedIDs == nil {
		state.WhatsAppProcessedIDs = make(map[string]int64)
	}
	if state.WhatsAppReplies == nil {
		state.WhatsAppReplies = make(map[string]MessageMapping)
	}
	if state.SignalReplies == nil {
		state.SignalReplies = make(map[string]MessageMapping)
	}

	// Migrate last_signal_timestamp from seconds to milliseconds if needed
	if state.LastSignalTimestamp < 100000000000 {
		state.LastSignalTimestamp = state.LastSignalTimestamp * 1000
	}
	return &state, nil
}

func SaveState(path string, state *State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(state)
}

func (s *State) CleanUpReplies() {
	cutoff := time.Now().Add(-7 * 24 * time.Hour).Unix()
	for id, mapping := range s.WhatsAppReplies {
		if mapping.Timestamp < cutoff {
			delete(s.WhatsAppReplies, id)
		}
	}
	for tsStr, mapping := range s.SignalReplies {
		if mapping.Timestamp < cutoff {
			delete(s.SignalReplies, tsStr)
		}
	}
}
