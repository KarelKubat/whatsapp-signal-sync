package main

import (
	"bufio"
	"cmp"
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow/types"
)

func RunSetup(ctx context.Context, cfgPath string, cfg *Config, waClient *WhatsAppClient, sigClient *SignalClient) error {
	fmt.Println("\n=============================================")
	fmt.Println("   WhatsApp - Signal Sync CLI: Group Setup   ")
	fmt.Println("=============================================\n")

	fmt.Println("Fetching WhatsApp joined groups...")
	waGroups, err := waClient.GetGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch WhatsApp groups: %v", err)
	}
	fmt.Printf("Found %d WhatsApp groups.\n", len(waGroups))

	fmt.Println("Fetching Signal joined groups...")
	sigGroups, err := sigClient.ListGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch Signal groups: %v", err)
	}
	fmt.Printf("Found %d Signal groups.\n\n", len(sigGroups))

	// Sort WhatsApp groups alphabetically (case-insensitive)
	slices.SortFunc(waGroups, func(a, b *types.GroupInfo) int {
		return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	// Sort Signal groups alphabetically (case-insensitive)
	slices.SortFunc(sigGroups, func(a, b SignalGroup) int {
		return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	if len(waGroups) == 0 {
		fmt.Println("No WhatsApp groups found to link. Setup complete.")
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		printWhatsAppGroups(waGroups, cfg, sigGroups)

		fmt.Print("Select WhatsApp group # to link/unlink (or 'd' if done): ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if strings.ToLower(input) == "d" {
			fmt.Println("Setup finalized.")
			break
		}

		waNum, err := strconv.Atoi(input)
		if err != nil || waNum < 1 || waNum > len(waGroups) {
			fmt.Printf("Invalid choice: %q. Please enter a number between 1 and %d, or 'd'.\n", input, len(waGroups))
			continue
		}

		selectedWA := waGroups[waNum-1]
		fmt.Printf("\nSelected WhatsApp Group: '%s'\n", selectedWA.Name)
		fmt.Printf("  JID: %s\n", selectedWA.JID.String())

		printSignalGroups(sigGroups)

		for {
			fmt.Printf("Enter Signal group # to link to (or 's' to unlink, 'n' to never mind): ")
			if !scanner.Scan() {
				break
			}
			sigInput := strings.TrimSpace(scanner.Text())
			if sigInput == "" {
				continue
			}

			if strings.ToLower(sigInput) == "n" {
				fmt.Println("Cancelled linking for this group.")
				break
			}

			if strings.ToLower(sigInput) == "s" {
				delete(cfg.GroupLinks, selectedWA.JID.String())
				fmt.Println("-> Unlinked WhatsApp group (set to personal forward).")
				break
			}

			sigNum, err := strconv.Atoi(sigInput)
			if err != nil || sigNum < 1 || sigNum > len(sigGroups) {
				fmt.Printf("Invalid choice: %q. Please enter a number between 1 and %d, 's' to unlink, or 'n'.\n", sigInput, len(sigGroups))
				continue
			}

			selectedSig := sigGroups[sigNum-1]
			cfg.GroupLinks[selectedWA.JID.String()] = selectedSig.ID
			fmt.Printf("-> Linked WhatsApp Group '%s' to Signal Group '%s'\n", selectedWA.Name, selectedSig.Name)
			break
		}
		fmt.Println()
	}

	fmt.Println("Saving configuration...")
	if err := SaveConfig(cfgPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}
	fmt.Println("Configuration saved successfully!")
	return nil
}

func printWhatsAppGroups(waGroups []*types.GroupInfo, cfg *Config, sigGroups []SignalGroup) {
	fmt.Println("\n--- Available WhatsApp Groups ---")
	for i, g := range waGroups {
		currentSigID, exists := cfg.GroupLinks[g.JID.String()]
		currentMappingName := "None (Personal Forward)"
		if exists {
			for _, sg := range sigGroups {
				if sg.ID == currentSigID {
					currentMappingName = sg.Name
					break
				}
			}
		}
		fmt.Printf("[%d] Name: %-30s | Current Link: %s\n", i+1, g.Name, currentMappingName)
	}
	fmt.Println("---------------------------------")
}

func printSignalGroups(sigGroups []SignalGroup) {
	fmt.Println("\n--- Available Signal Groups ---")
	for i, g := range sigGroups {
		fmt.Printf("[%d] Name: %s (ID: %s)\n", i+1, g.Name, g.ID)
	}
	fmt.Println("-------------------------------")
}
