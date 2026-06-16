package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	_ "github.com/mattn/go-sqlite3"
)

type WhatsAppClient struct {
	dbPath         string
	client         *whatsmeow.Client
	db             *sqlstore.Container
	incomingEvents chan *events.Message
	done           chan struct{}
}

func NewWhatsAppClient(dbPath string) *WhatsAppClient {
	return &WhatsAppClient{
		dbPath:         dbPath,
		incomingEvents: make(chan *events.Message, 100),
		done:           make(chan struct{}),
	}
}

func (w *WhatsAppClient) Start(ctx context.Context) error {
	// Ensure parent directory for database exists
	if err := os.MkdirAll(filepath.Dir(w.dbPath), 0700); err != nil {
		return err
	}

	// whatsmeow logs can be verbose, use an error logger to silence harmless warnings on shutdown
	logLevel := "ERROR"
	dbLog := waLog.Stdout("Database", logLevel, true)
	clientLog := waLog.Stdout("WhatsAppClient", logLevel, true)

	var err error
	w.db, err = sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", w.dbPath), dbLog)
	if err != nil {
		return err
	}

	deviceStore, err := w.db.GetFirstDevice(ctx)
	if err != nil {
		return err
	}

	w.client = whatsmeow.NewClient(deviceStore, clientLog)
	w.client.AddEventHandler(w.eventHandler)

	if w.client.Store.ID == nil {
		// No logged-in session, perform login flow
		qrChan, err := w.client.GetQRChannel(ctx)
		if err != nil {
			return err
		}

		err = w.client.Connect()
		if err != nil {
			return err
		}

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case qrCode, ok := <-qrChan:
				if !ok {
					// Channel closed, indicating registration completed or timed out
					break
				}
				if qrCode.Event == "code" {
					w.renderQR(qrCode.Code)
				}
			}
			if w.client.IsLoggedIn() {
				break
			}
		}
	} else {
		// Logged in session exists, just connect
		err = w.client.Connect()
		if err != nil {
			return err
		}
	}

	// Wait to make sure connection is fully synchronized
	time.Sleep(2 * time.Second)
	return nil
}

func (w *WhatsAppClient) Stop() {
	close(w.done)
	if w.client != nil {
		w.client.Disconnect()
	}
}

func (w *WhatsAppClient) renderQR(code string) {
	fmt.Println("\n--- Scan the QR code below with your WhatsApp app ---")
	qr, err := qrcode.New(code, qrcode.Medium)
	if err != nil {
		log.Printf("Failed to generate WhatsApp QR code: %v", err)
		return
	}
	fmt.Println(qr.ToSmallString(false))
	fmt.Println("------------------------------------------------------\n")
}

func (w *WhatsAppClient) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		w.incomingEvents <- v
	}
}

func (w *WhatsAppClient) SendTextMessage(ctx context.Context, to JID, text string) (string, error) {
	var targetJID types.JID
	var err error
	if to.IsGroup {
		targetJID, err = types.ParseJID(to.String())
	} else {
		targetJID, err = types.ParseJID(to.String())
	}
	if err != nil {
		return "", err
	}

	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	resp, err := w.client.SendMessage(ctx, targetJID, msg)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (w *WhatsAppClient) SendMediaMessage(ctx context.Context, to JID, data []byte, mimeType string, isVideo bool, caption string) (string, error) {
	targetJID, err := types.ParseJID(to.String())
	if err != nil {
		return "", err
	}

	var mediaType whatsmeow.MediaType
	if isVideo {
		mediaType = whatsmeow.MediaVideo
	} else {
		mediaType = whatsmeow.MediaImage
	}

	uploadResp, err := w.client.Upload(ctx, data, mediaType)
	if err != nil {
		return "", err
	}

	var msg *waE2E.Message
	if isVideo {
		msg = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(data))),
				Caption:       proto.String(caption),
			},
		}
	} else {
		msg = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(data))),
				Caption:       proto.String(caption),
			},
		}
	}

	resp, err := w.client.SendMessage(ctx, targetJID, msg)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (w *WhatsAppClient) DownloadMedia(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
	return w.client.Download(ctx, msg)
}

func (w *WhatsAppClient) GetGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	return w.client.GetJoinedGroups(ctx)
}

// Simple wrapper around JID string to distinguish user / group JIDs cleanly
type JID struct {
	Raw     string
	IsGroup bool
}

func (j JID) String() string {
	return j.Raw
}

func ParseWhatsAppJID(s string) (JID, error) {
	jid, err := types.ParseJID(s)
	if err != nil {
		return JID{}, err
	}
	return JID{
		Raw:     jid.String(),
		IsGroup: jid.Server == types.GroupServer,
	}, nil
}
