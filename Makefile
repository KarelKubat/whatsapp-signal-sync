me a sandwich:
	GOOS=darwin GOARCH=arm64 go build -o whatsapp-signal-sync-darwin-arm64    # ARM-based Mac
	GOOS=linux GOARCH=arm GOARM=5 go build -o whatsapp-signal-sync-linux-arm5 # Raspberry Pi
	GOOS=linux GOARCH=amd64 go build -o whatsapp-signal-sync-linux-amd64      # Linux x86_64
