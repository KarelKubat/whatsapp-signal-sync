# Testing

When running `whatsapp-signal-sync -setup` I see:

```
2026/06/16 16:21:22 Starting WhatsApp-Signal Sync CLI...
2026/06/16 16:21:23 Connecting to WhatsApp...
2026/06/16 16:21:25 Connected to WhatsApp.
2026/06/16 16:21:25 Connecting to Signal daemon...
2026/06/16 16:21:25 Connected to Signal.

=============================================
   WhatsApp - Signal Sync CLI: Group Setup
=============================================

Fetching WhatsApp joined groups...
Found 72 WhatsApp groups.
Fetching Signal joined groups...
Found 1 Signal groups.

--- Available Signal Groups ---
[1] Name: Eexter Meiborgen (ID: aO/uf3Uk+H10SHqNE0iQZqSSQrqASOJHadr2iX/o+Pw=)
-------------------------------

WhatsApp Group: 'House chat ☺️'
  JID: 120363041400886985@g.us
  Current Link: None (Personal Forward)
Link to Signal group # (or 's' to skip/for personal forward, 'd' if done): d
Setup finalized.
Saving configuration...
Configuration saved successfully!
16:21:36.234 [WhatsAppClient/Socket WARN] Error sending close to websocket: failed to close WebSocket: failed to read frame header: EOF
```

Fix the problem.
