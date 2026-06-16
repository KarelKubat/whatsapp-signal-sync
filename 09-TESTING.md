# Testing

While running `whatsapp-signal-sync -setup` I get past scanning the Signal and WhatsApp QR codes, but then the output shows:

```sh
=============================================
   WhatsApp - Signal Sync CLI: Group Setup
=============================================

Fetching WhatsApp joined groups...
Found 72 WhatsApp groups.
Fetching Signal joined groups...
2026/06/16 16:18:32 [signal-cli stderr] INFO  AccountHelper - The Signal protocol expects that incoming messages are regularly received.
16:18:32.821 [Database WARN] 439 duplicate contacts found in PutAllContactNames
2026/06/16 16:18:33 Setup failed: failed to fetch Signal groups: json: cannot unmarshal object into Go struct field SignalGroup.members of type string
```

Fix the problem.
