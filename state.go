package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	LastWhatsAppTimestamp int64  `json:"last_whatsapp_timestamp"`
	LastWhatsAppMsgID     string `json:"last_whatsapp_msg_id"`
	LastSignalTimestamp   int64  `json:"last_signal_timestamp"`
}

func LoadState(path string) (*State, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		// First run: default to current time to avoid importing ancient history
		now := time.Now().Unix()
		state := &State{
			LastWhatsAppTimestamp: now,
			LastSignalTimestamp:   now * 1000,
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
