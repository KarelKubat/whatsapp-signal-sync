# whatsapp-signal-sync

A self-contained CLI utility designed to automatically synchronize your personal and group messages, images, and videos in real time between WhatsApp and Signal.

---

## 1. What is it, what does it do?

`whatsapp-signal-sync` is a bridging program written in Go that acts as a real-time forwarder between your WhatsApp and Signal accounts. 

### But why?

![But Why](butwhy.png)

I am member of a WhatsApp family group where lots of people hang out, despite that many are not happy about it. There is just too much inertial mass to switch to Signal. So most of the members end up being on WhatsApp "just because of that group". 

Some people however strongly prefer Signal - so - would it not be nice to have two worlds? Group members on the one platform would see messages on the other platform too, and vice versa.

And thus the idea for `whatsapp-signal-sync` was born. It allows the folks who want to be on Signal (who are mostly more tech-savvy anyway) to run a syncer on their favorite Linux box, or on a Raspberry Pi, or on a little Mac mini. The less tech-savvy can stay on WhatsApp. With this setup, no one has to sacrifice their preferred platform.

### Key Features:
- **Two-way Syncing**: Messages sent or received on WhatsApp are forwarded to Signal, and vice versa.
- **Group linking**: Connects specific WhatsApp groups to corresponding Signal groups to mirror group messages.
- **Message Redirection & Replies**:
  - Direct and group messages are automatically synced between linked accounts/groups.
  - **Replying (quoting)** a forwarded message in your chat on either Signal or WhatsApp automatically routes your reply back to the correct recipient or mapped group on the opposite platform.
  - **Unlinked Group Safety**: If you reply to a message from an unlinked group, the syncer blocks the reply from being sent to the other group, and instead redirects it to your personal account on the other side prefixed with a `[Group Reply Failed]` warning.
- **Clean Signal Group Integration**: If you create a Signal group named `"Whatsapp Signal Sync"`, the program automatically detects it and forwards all direct/personal and unlinked group messages there instead of polluting your personal `Note to Self` chat.
- **Media Support**: Automatically downloads and transfers images and videos in transit.
- **Historical Catch-up Sync**: Remembers where it left off by saving its synchronization state in `./data/state.json`. If the daemon is stopped or suffers a network blackout, it automatically pulls and syncs any missed messages upon startup and re-triggers polling every 5 minutes.
- **Safe Fallbacks**: Unsupported messages (like polls) are replaced with a clear text placeholder notifying you to check the original message on the source platform.

---

## 2. Onboarding & Usage (End-User Guide)

Once `whatsapp-signal-sync` is installed, follow these steps to link your accounts and start syncing.

### Step 2.1: Initial Configuration & Registration
Run the interactive setup wizard by executing:
```bash
whatsapp-signal-sync -setup
```

1. **Enter Your Accounts Info**:
   - The program will prompt you to enter your **Signal Phone Number** (e.g. `+1234567890`) and your **WhatsApp JID** (e.g. `1234567890@s.whatsapp.net` which is your WhatsApp phone number followed by `@s.whatsapp.net`).
2. **Link Signal**:
   - The program will output a **Signal QR code** in your terminal.
   - Open your Signal app on your phone, go to **Settings -> Linked Devices -> Link New Device**, and scan this QR code.
3. **Link WhatsApp**:
   - The program will then output a **WhatsApp QR code** in your terminal.
   - Open your WhatsApp app on your phone, go to **Linked Devices -> Link a Device**, and scan this QR code.

### Step 2.2: Interactive Group Linking
- After both accounts are linked, the program will fetch all groups you belong to on both WhatsApp and Signal.
- It will go through each WhatsApp group one-by-one and ask you to enter the number of the corresponding Signal group to link them together.
- Enter the number of the matching Signal group, or press `s` to skip mapping (unlinked group messages will default to personal forwarding), or `d` if you are done.
- The mapping is saved to `./data/config.yaml`.

You can re-link groups anytime by re-running `whatsapp-signal-sync -setup`. The program will not ask you again to re-login. If you want to re-login, you have to delete the ./data/whatsapp.db file.

### Step 2.3: Running the Daemon
To start the real-time sync engine, run the program without flags:
```bash
whatsapp-signal-sync
```
Keep this process running in your terminal (or run it under a process manager like systemd, launchd, or screen). To stop the sync daemon at any time, press `Ctrl+C`.

---

## 3. Compilation & Installation (Sysadmin Guide)

Follow these instructions to compile the binary and set up the execution environment from source.

### Prerequisites:
1. **Go Toolchain**: Go 1.25+ or 1.26+ must be installed.
2. **signal-cli**: The binary `signal-cli` (v0.10.0 or higher) must be installed and available in the system path (`$PATH`).
   - On macOS: `brew install signal-cli`
   - On Linux: Download and install the latest tarball from the [signal-cli release page](https://github.com/AsamK/signal-cli/releases).

### Compilation Steps:

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/KarelKubat/whatsapp-signal-sync.git
   cd whatsapp-signal-sync
   ```

2. **Download Dependencies**:
   ```bash
   go mod download
   ```

3. **Build the Executable**:
   ```bash
   go build -o whatsapp-signal-sync .
   ```

4. **Verify Compilation**:
   ```bash
   ./whatsapp-signal-sync --help
   ```

### Operational Directories & Configuration
By default, the program creates and uses the following structure:
- `./data/config.yaml`: Contains configuration parameters and group links.
- `./data/whatsapp.db`: SQLite database storing active WhatsApp sessions.
- `./data/signal/`: Holds Signal device keys, profiles, and configuration (used by `signal-cli`).
- `./data/tmp/`: Used for temporary media caching during image/video forwarding.

Example of `./data/config.yaml`:
```yaml
storage:
  whatsapp_db: "./data/whatsapp.db"
  signal_config_dir: "./data/signal"
  temp_attachment_dir: "./data/tmp"
  signal_cli_path: "signal-cli" # Path to signal-cli (supports absolute paths)

accounts:
  signal_number: "+1234567890"
  whatsapp_user_jid: "1234567890@s.whatsapp.net"

group_links:
  "120363024888888888@g.us": "EdY4T25/Tf+1Hk8gY1/p5Q=="
```
