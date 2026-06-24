#!/bin/sh

make || exit 1

OS=$(uname -s)
ARCH=$(uname -m)
BINARY=""

case "$OS" in
    Darwin)
        case "$ARCH" in
            arm64) BINARY="./whatsapp-signal-sync-darwin-arm64" ;;
        esac
        ;;
    Linux)
        case "$ARCH" in
            x86_64|amd64) BINARY="./whatsapp-signal-sync-linux-amd64" ;;
            arm*|aarch64) BINARY="./whatsapp-signal-sync-linux-arm5" ;;
        esac
        ;;
esac

if [ -z "$BINARY" ]; then
    echo "Unsupported platform: $OS $ARCH" >&2
    exit 1
fi

"$BINARY" -debug 2>&1 | ./loglimit.pl /tmp/whatsapp-signal-sync.log 100000 5
