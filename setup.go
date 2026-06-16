package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
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

	if len(waGroups) == 0 {
		fmt.Println("No WhatsApp groups found to link. Setup complete.")
		return nil
	}

	printSignalGroups(sigGroups)

	scanner := bufio.NewScanner(os.Stdin)

	for _, waGrp := range waGroups {
		// Display current mapping if it exists
		currentSigID, exists := cfg.GroupLinks[waGrp.JID.String()]
		currentMappingName := "None (Personal Forward)"
		if exists {
			for _, g := range sigGroups {
				if g.ID == currentSigID {
					currentMappingName = g.Name
					break
				}
			}
		}

		fmt.Printf("WhatsApp Group: '%s'\n", waGrp.Name)
		fmt.Printf("  JID: %s\n", waGrp.JID.String())
		fmt.Printf("  Current Link: %s\n", currentMappingName)

		for {
			fmt.Printf("Link to Signal group # (or 's' to skip, 'l' to list, 'd' if done): ")
			if !scanner.Scan() {
				break
			}
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			if strings.ToLower(input) == "s" {
				delete(cfg.GroupLinks, waGrp.JID.String())
				fmt.Println("-> Unlinked / set to personal forward.")
				break
			}

			if strings.ToLower(input) == "l" {
				printSignalGroups(sigGroups)
				continue
			}

			if strings.ToLower(input) == "d" {
				fmt.Println("Setup finalized.")
				goto SaveAndExit
			}

			num, err := strconv.Atoi(input)
			if err != nil || num < 1 || num > len(sigGroups) {
				fmt.Printf("Invalid choice: %q. Please enter a number between 1 and %d.\n", input, len(sigGroups))
				continue
			}

			selectedSig := sigGroups[num-1]
			cfg.GroupLinks[waGrp.JID.String()] = selectedSig.ID
			fmt.Printf("-> Linked to Signal Group: '%s' (ID: %s)\n", selectedSig.Name, selectedSig.ID)
			break
		}
		fmt.Println()
	}

SaveAndExit:
	fmt.Println("Saving configuration...")
	if err := SaveConfig(cfgPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}
	fmt.Println("Configuration saved successfully!")
	return nil
}

func printSignalGroups(sigGroups []SignalGroup) {
	fmt.Println("\n--- Available Signal Groups ---")
	for i, g := range sigGroups {
		fmt.Printf("[%d] Name: %s (ID: %s)\n", i+1, g.Name, g.ID)
	}
	fmt.Println("-------------------------------\n")
}
