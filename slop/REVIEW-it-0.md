# Code Review - WhatsApp-Signal Sync CLI (Iteration 0)

This document presents a code review of the `whatsapp-signal-sync` CLI codebase. The review assesses code structure, Golang best practices, modularization, security, and usability.

---

## 1. Modularization & Package Structure

### Findings:
- **Single Package (`package main`)**: All Go files are declared under `package main`. For a small CLI program, this is acceptable, but it leads to global namespace pollution and hampers unit testing.
- **Tight Coupling**: The `SyncEngine` directly coordinates concrete client implementations (`WhatsAppClient` and `SignalClient`) rather than using Go interfaces.

### Recommendations:
- **Extract Interfaces**: Define `MessageSender`, `MessageReceiver`, and `MediaDownloader` interfaces. This will decouple the Sync Engine from the concrete WhatsApp and Signal implementations, allowing you to mock them for unit tests.
- **Restructure into Subpackages**: Consider moving platform-specific logic into subdirectories:
  - `pkg/config/`
  - `pkg/whatsapp/`
  - `pkg/signal/`
  - `pkg/sync/`

---

## 2. Subprocess Management & Reliability

### Findings:
- **Subprocess Failures**: In `signal.go`, if the `signal-cli` JSON-RPC daemon subprocess crashes or exits due to a connection drop or authentication failure, the read loop terminates, but the main Go program does not notice or attempt to restart it. The program will simply go silent.
- **Buffered Scanner Overflow**: The stdout read loop uses `bufio.NewScanner` with default buffer size (max 64KB). Extremely large JSON-RPC responses (e.g. from `listGroups` with hundreds of groups or members) might overflow the scanner and crash the loop.
- **Hardcoded Sleep in WhatsApp Startup**: In `whatsapp.go`, `time.Sleep(2 * time.Second)` is used after connection to wait for synchronization. Naive sleeps can be flaky depending on network latency.

### Recommendations:
- **Subprocess Supervision**: Wrap the `signal-cli` command execution in a restart loop with exponential backoff. If the process terminates unexpectedly, log the event and restart the daemon.
- **Custom Scanner Buffer**: Set a larger buffer size for the stdout scanner (e.g. up to 1MB) or use a raw line-by-line reader (`bufio.Reader.ReadLine` or `ReadSlice`) to prevent scanner overflows.
- **Listen to whatsmeow Events**: Replace the `time.Sleep` in `whatsapp.go` by listening to the connection events (`*events.Connected` / `*events.LoggedOut`) to determine client status dynamically.

---

## 3. Security Analysis

### Findings:
- **Permissive File Permissions**: Configurations and databases are created with standard permissive flags (directory: `0755`, file: `0644`). Because `config.yaml` contains phone numbers, group identifiers, and potentially tokens, and `whatsapp.db` contains session cryptographic tokens, this represents a security concern.
- **System Path Dependency**: The binary directly invokes `exec.Command("signal-cli", ...)` depending on the system's global `$PATH`. If a malicious binary named `signal-cli` is placed in the user's path, it could lead to arbitrary binary execution.
- **Lingering Temporary Files**: If the daemon crashes while forwarding a media file, the staged media files in `temp_attachment_dir` remain on disk indefinitely.

### Recommendations:
- **Tighten File Permissions**: Use `0600` permissions (read/write only by owner) when creating files like `config.yaml` and databases, and `0700` for directory creation.
- **Configurable Binary Path**: Add a `signal_cli_path` configuration parameter to the YAML file, defaulting to `"signal-cli"`. This lets sysadmins pin the exact absolute path to the official executable (e.g. `/opt/homebrew/bin/signal-cli`).
- **Temporary Folder Cleanup**: Add a startup cleanup task that purges any left-over files in the `temp_attachment_dir` when the program launches.

---

## 4. Golang Best Practices

### Findings:
- **Context Propagation**: Standard Go context propagation is implemented nicely for WhatsApp and config operations. However, in `signal.go`, the JSON-RPC stdin write does not respect context timeouts, only the response channel read does.
- **Structured Logging**: Standard `log.Println` is used throughout the sync engine. While simple, it lacks levels (DEBUG, INFO, ERROR) and JSON structure.

### Recommendations:
- **Respect Context in Stdin Writes**: Use non-blocking channel checks or select statements on `ctx.Done()` when writing to `s.stdin` to avoid blocking when the subprocess standard input is stalled.
- **Leverage Zerolog**: Since `github.com/rs/zerolog` is already downloaded as a dependency of `whatsmeow`/`mautrix`, leverage it for unified, structured logging across both WhatsApp, Signal, and the sync engine.
