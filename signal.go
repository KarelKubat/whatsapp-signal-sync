package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      string      `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      string          `json:"id"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SignalMessageEvent struct {
	Envelope struct {
		Source       string `json:"source"`
		SourceName   string `json:"sourceName"`
		SourceNumber string `json:"sourceNumber"`
		SourceUUID   string `json:"sourceUuid"`
		Timestamp    int64  `json:"timestamp"`
		DataMessage  *struct {
			Timestamp   int64  `json:"timestamp"`
			Message     string `json:"message"`
			Quote       *struct {
				ID     int64  `json:"id"`
				Author string `json:"author"`
				Text   string `json:"text"`
			} `json:"quote"`
			GroupInfo   *struct {
				GroupID string              `json:"groupId"`
				Members []SignalGroupMember `json:"members"`
				Name    string              `json:"name"`
				Type    string              `json:"type"`
			} `json:"groupInfo"`
			Attachments []struct {
				ContentType    string `json:"contentType"`
				Filename       string `json:"filename"`
				ID             string `json:"id"`
				Size           int64  `json:"size"`
				StoredFilename string `json:"storedFilename"`
			} `json:"attachments"`
		} `json:"dataMessage"`
	} `json:"envelope"`
	Account string `json:"account"`
}

type SignalGroupMember struct {
	Number  *string `json:"number"`
	UUID    string  `json:"uuid"`
	IsAdmin bool    `json:"isAdmin"`
}

type SignalGroup struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Members []SignalGroupMember `json:"members"`
	Blocked bool                `json:"blocked"`
}

type SignalClient struct {
	signalCLIPath string
	configDir     string
	account       string
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser

	mu              sync.Mutex
	pendingRequests map[string]chan *JSONRPCResponse
	incomingEvents  chan *SignalMessageEvent
	done            chan struct{}
}

func NewSignalClient(signalCLIPath, configDir, account string) *SignalClient {
	return &SignalClient{
		signalCLIPath:   signalCLIPath,
		configDir:       configDir,
		account:         account,
		pendingRequests: make(map[string]chan *JSONRPCResponse),
		incomingEvents:  make(chan *SignalMessageEvent, 100),
		done:            make(chan struct{}),
	}
}

func (s *SignalClient) Start() error {
	if err := s.startOnce(); err != nil {
		return err
	}
	go s.supervisor()
	return nil
}

func (s *SignalClient) startOnce() error {
	args := []string{}
	if s.configDir != "" {
		args = append(args, "--config", s.configDir)
	}
	if s.account != "" {
		args = append(args, "--account", s.account)
	}
	args = append(args, "jsonRpc")

	s.cmd = exec.Command(s.signalCLIPath, args...)

	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return err
	}

	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start log wrapper for stderr so we don't lose debugging context
	go func() {
		scanner := bufio.NewScanner(s.stderr)
		for scanner.Scan() {
			log.Printf("[signal-cli stderr] %s", scanner.Text())
		}
	}()

	if err := s.cmd.Start(); err != nil {
		return err
	}

	go s.readLoop(s.stdout)

	return nil
}

func (s *SignalClient) Stop() error {
	close(s.done)
	if s.stdin != nil {
		s.stdin.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		return s.cmd.Process.Kill()
	}
	return nil
}

func (s *SignalClient) supervisor() {
	for {
		// Wait for command to finish
		err := s.cmd.Wait()

		select {
		case <-s.done:
			// Client stopped intentionally
			return
		default:
			log.Printf("Signal subprocess exited unexpectedly with error: %v. Restarting in 5 seconds...", err)
			s.clearPendingRequests(err)
			time.Sleep(5 * time.Second)

			// Re-connect
			if err := s.startOnce(); err != nil {
				log.Printf("Failed to restart Signal subprocess: %v. Retrying in 10 seconds...", err)
				time.Sleep(10 * time.Second)
			} else {
				log.Println("Signal subprocess restarted successfully.")
			}
		}
	}
}

func (s *SignalClient) clearPendingRequests(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for reqID, ch := range s.pendingRequests {
		ch <- &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      reqID,
			Error: &JSONRPCError{
				Code:    -32603,
				Message: fmt.Sprintf("connection lost: %v", err),
			},
		}
		delete(s.pendingRequests, reqID)
	}
}

func (s *SignalClient) readLoop(stdout io.ReadCloser) {
	scanner := bufio.NewScanner(stdout)
	// Set 1MB scanner buffer limit to handle large payloads (e.g. lists of groups)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Try parsing as response first
		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err == nil && resp.ID != "" {
			s.mu.Lock()
			ch, ok := s.pendingRequests[resp.ID]
			if ok {
				ch <- &resp
				delete(s.pendingRequests, resp.ID)
			}
			s.mu.Unlock()
			continue
		}

		// Try parsing as incoming event (notification)
		var raw map[string]interface{}
		if err := json.Unmarshal(line, &raw); err == nil {
			if method, ok := raw["method"].(string); ok && method == "receive" {
				var event SignalMessageEvent
				if err := json.Unmarshal(line, &event); err == nil {
					s.incomingEvents <- &event
				} else {
					log.Printf("Failed to unmarshal Signal message event: %v", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Signal client stdout scanner error: %v", err)
	}
}

func (s *SignalClient) sendRequest(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	reqID := uuid.New().String()
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      reqID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	ch := make(chan *JSONRPCResponse, 1)
	s.mu.Lock()
	s.pendingRequests[reqID] = ch
	s.mu.Unlock()

	s.mu.Lock()
	_, err = s.stdin.Write(data)
	s.mu.Unlock()
	if err != nil {
		s.mu.Lock()
		delete(s.pendingRequests, reqID)
		s.mu.Unlock()
		return nil, err
	}

	select {
	case <-ctx.Done():
		s.mu.Lock()
		delete(s.pendingRequests, reqID)
		s.mu.Unlock()
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("signal-cli error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

func (s *SignalClient) ListGroups(ctx context.Context) ([]SignalGroup, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	result, err := s.sendRequest(ctx, "listGroups", nil)
	if err != nil {
		return nil, err
	}

	var groups []SignalGroup
	if err := json.Unmarshal(result, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

type SendParams struct {
	Message     string   `json:"message,omitempty"`
	Recipient   string   `json:"recipient,omitempty"`
	Group       string   `json:"group,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

func (s *SignalClient) SendMessage(ctx context.Context, recipient, group, message string, attachments []string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := SendParams{
		Message:     message,
		Attachments: attachments,
	}
	if group != "" {
		params.Group = group
	} else if recipient != "" {
		params.Recipient = recipient
	} else {
		return errors.New("must specify either recipient or group")
	}

	_, err := s.sendRequest(ctx, "send", params)
	return err
}

// Helper to Link Device as secondary device and return the URI QR code channel
func LinkDevice(ctx context.Context, signalCLIPath, configDir, deviceName string, qrChan chan<- string) error {
	args := []string{}
	if configDir != "" {
		args = append(args, "--config", configDir)
	}
	args = append(args, "link")
	if deviceName != "" {
		args = append(args, "-n", deviceName)
	}

	cmd := exec.Command(signalCLIPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[signal-cli link stderr] %s", scanner.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) > 0 {
				qrChan <- line
			}
		}
	}()

	// Wait for command completion
	errChan := make(chan error, 1)
	go func() {
		errChan <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
