# Testing

I am running `whatsapp-signal-sync -setup`. The output on screen is:

```sh
2026/06/16 16:11:26 Starting WhatsApp-Signal Sync CLI...
Signal account is not linked. Starting linking process...
2026/06/16 16:12:27 [signal-cli link stderr] Link request error: Connection closed!
2026/06/16 16:12:27 Signal linking failed: exit status 3
```

Meanwhile I did see in the process list:

```sh
signal-cli --config ./data/signal link -n whatsapp-sync
```

Something is wrong:

1. When I run by hand `signal-cli --config ./data/signal link -n whatsapp-sync` then I see a QR code on screen that I could scan.
2. When I run  `whatsapp-signal-sync -setup` then I assume that this runs the underlying `signal-cli --config ./data/signal link -n whatsapp-sync`. But nothing appears on screen and it times out.

Fix the problem.
