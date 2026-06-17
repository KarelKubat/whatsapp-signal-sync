package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/skip2/go-qrcode"
)

func main() {
	configPath := flag.String("config", "./data/config.yaml", "Path to YAML configuration file")
	setupMode := flag.Bool("setup", false, "Run interactive group linking setup and exit")
	debugMode := flag.Bool("debug", false, "Run in debug mode (verbose logging and headers)")
	flag.Parse()

	log.Println("Starting WhatsApp-Signal Sync CLI...")

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	cfg.Debug = *debugMode

	// Clean up temporary attachment directory on startup
	if cfg.Storage.TempAttachmentDir != "" {
		if err := clearTempAttachmentDir(cfg.Storage.TempAttachmentDir); err != nil {
			log.Printf("Warning: Failed to clean up temp attachment directory on startup: %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// If accounts are not configured, perform initial configuration
	if cfg.Accounts.SignalNumber == "" || cfg.Accounts.WhatsAppUserJID == "" {
		fmt.Println("\n--- Initial Account Configuration ---")
		scanner := bufio.NewScanner(os.Stdin)

		if cfg.Accounts.SignalNumber == "" {
			fmt.Print("Enter your Signal Phone Number (e.g. +1234567890): ")
			if scanner.Scan() {
				cfg.Accounts.SignalNumber = strings.TrimSpace(scanner.Text())
			}
		}

		if cfg.Accounts.WhatsAppUserJID == "" {
			fmt.Print("Enter your WhatsApp User JID (e.g. 1234567890@s.whatsapp.net): ")
			if scanner.Scan() {
				cfg.Accounts.WhatsAppUserJID = strings.TrimSpace(scanner.Text())
			}
		}

		if err := SaveConfig(*configPath, cfg); err != nil {
			log.Fatalf("Failed to save initial configuration: %v", err)
		}
		fmt.Println("Initial configuration saved successfully.\n")
	}

	// Verify Signal linking
	isSignalLinked, err := checkIfSignalLinked(cfg.Storage.SignalCLIPath, cfg.Storage.SignalConfigDir, cfg.Accounts.SignalNumber)
	if err != nil {
		log.Fatalf("Failed to check Signal account link status: %v", err)
	}

	if !isSignalLinked {
		fmt.Println("Signal account is not linked. Starting linking process...")
		qrChan := make(chan string, 1)

		go func() {
			for linkURI := range qrChan {
				if strings.HasPrefix(linkURI, "tsdevice:") || strings.HasPrefix(linkURI, "sgnl://") {
					fmt.Println("\n--- Scan the QR code below with your Signal app (Settings -> Linked Devices) ---")
					qr, err := qrcode.New(linkURI, qrcode.Medium)
					if err == nil {
						fmt.Println(qr.ToSmallString(false))
					} else {
						fmt.Printf("Link URI: %s\n", linkURI)
					}
					fmt.Println("---------------------------------------------------------------------------------\n")
				}
			}
		}()

		err := LinkDevice(ctx, cfg.Storage.SignalCLIPath, cfg.Storage.SignalConfigDir, "whatsapp-sync", qrChan)
		close(qrChan)
		if err != nil {
			log.Fatalf("Signal linking failed: %v", err)
		}
		fmt.Println("Signal linking completed successfully!")
	}

	// Initialize Clients
	if cfg.Debug {
		log.Println("[DEBUG] Initializing WhatsApp and Signal clients...")
	}
	waClient := NewWhatsAppClient(cfg.Storage.WhatsAppDB)
	waClient.Debug = cfg.Debug
	sigClient := NewSignalClient(cfg.Storage.SignalCLIPath, cfg.Storage.SignalConfigDir, cfg.Accounts.SignalNumber)
	sigClient.Debug = cfg.Debug

	// Start WhatsApp Client
	if cfg.Debug {
		log.Println("[DEBUG] Connecting to WhatsApp (checking DB and device store)...")
	} else {
		log.Println("Connecting to WhatsApp...")
	}
	if err := waClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start WhatsApp client: %v", err)
	}
	defer waClient.Stop()
	if cfg.Debug {
		log.Println("[DEBUG] WhatsApp connection established successfully.")
	} else {
		log.Println("Connected to WhatsApp.")
	}

	// Start Signal Client
	if cfg.Debug {
		log.Println("[DEBUG] Spawning Signal daemon subprocess with JSON-RPC...")
	} else {
		log.Println("Connecting to Signal daemon...")
	}
	if err := sigClient.Start(); err != nil {
		log.Fatalf("Failed to start Signal client: %v", err)
	}
	defer func() {
		_ = sigClient.Stop()
	}()
	if cfg.Debug {
		log.Println("[DEBUG] Connected to Signal daemon successfully.")
	} else {
		log.Println("Connected to Signal.")
	}

	if *setupMode {
		// Run setup linking mode and exit
		err := RunSetup(ctx, *configPath, cfg, waClient, sigClient)
		if err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		return
	}

	// Start Sync Engine
	if cfg.Debug {
		log.Println("[DEBUG] Initializing synchronization engine and starting listener routines...")
	} else {
		log.Println("Starting synchronization engine...")
	}
	engine := NewSyncEngine(cfg, waClient, sigClient)
	engine.Start(ctx)
	defer engine.Stop()
	if cfg.Debug {
		log.Println("[DEBUG] Synchronization loops started. Listening for WhatsApp events and Signal socket notifications.")
	}
	log.Println("Synchronization engine is running. Press Ctrl+C to stop.")

	// Capture interrupt signals for graceful termination
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	<-shutdownChan

	log.Println("Shutting down clients and exiting...")
}

func checkIfSignalLinked(signalCLIPath, configDir, number string) (bool, error) {
	args := []string{}
	if configDir != "" {
		args = append(args, "--config", configDir)
	}
	args = append(args, "listAccounts")

	cmd := exec.Command(signalCLIPath, args...)
	output, err := cmd.Output()
	if err != nil {
		// If command fails because directory is empty or no accounts, it's not linked
		return false, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, number) {
			return true, nil
		}
	}
	return false, nil
}

func clearTempAttachmentDir(dir string) error {
	d, err := os.Open(dir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			log.Printf("Failed to remove temp file %s: %v", name, err)
		}
	}
	return nil
}
