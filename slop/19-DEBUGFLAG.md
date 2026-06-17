# Debug flag

Add a bool flag `--debug` with the default `false`. When set to `true`, the syncer should:

- Log in which state the syncer is (e.g., starting up the clients, parsing the configs, listening)
- Log all incoming and outgoing messages with full payload
- Show headers in the messages on WhatsApp and Signal, but suppress them when `--debug` is `false`. What I mean with that is described below.

The current code prefixes forwarded messages with a header, e.g. `[WhatsApp Direct: 1234567890@lid]`. This is then followed by the name of the original sender and with the message content. The header between `[` and `]` doesn't look great. The forwarded message is readable just fine when it has only the name and the message text.

Is it possible to suppress the header in normal, non-debug operation? If yes, do so. If not, you may leave it as-is (because e.g. it is needed to determine that the message was handled by the syncer).
