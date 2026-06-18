me a sandwich:
	GOOS=darwin GOARCH=arm64 go build                           # for an ARM-based Mac
	mv whatsapp-signal-sync whatsapp-signal-sync-darwin-arm64
	GOOS=linux GOARCH=arm GOARM=5 go build                      # for Raspberry Pi
	mv whatsapp-signal-sync whatsapp-signal-sync-linux-arm5
