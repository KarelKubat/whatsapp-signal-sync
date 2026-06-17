# Testing

I see the following output:

```
2026/06/16 19:13:17 Starting WhatsApp-Signal Sync CLI...
2026/06/16 19:13:19 Connecting to WhatsApp...
2026/06/16 19:13:21 Connected to WhatsApp.
2026/06/16 19:13:21 Connecting to Signal daemon...
2026/06/16 19:13:21 Connected to Signal.
2026/06/16 19:13:21 Starting synchronization engine...
2026/06/16 19:13:23 [SyncEngine] No 'Whatsapp Signal Sync' Signal group found. Defaulting personal forwards to 'Note to Self' (your Signal number).
2026/06/16 19:13:23 [SyncEngine] Loaded state: WhatsApp Last Timestamp: 1781630003, Signal Last Timestamp: 1781630003
2026/06/16 19:13:23 Synchronization engine is running. Press Ctrl+C to stop.
2026/06/16 19:13:23 [SyncEngine] Triggering initial Signal receive catch-up...
2026/06/16 19:13:23 [SyncEngine] Initial Signal receive trigger failed: signal-cli error -1: Receive command cannot be used if messages are already being received.
```

Check whether the last line is critical. If it is not then you can leave it as-is. If it harms, fix the problem.
