# Catching up old messages

Modify the syncer so that missed messages are synced. A usecase is e.g. when the syncer was stopped, and is now started. It should look back and sync missed messages.

The syncer must not only do this upon startup, but also periodically, say each 5 minutes. Then, incase there was a network outage, older messages are still synced.

So this means that the syncer must become state-aware. I suggest to let the state depend on a timestamp, but you can implement something else (sequence number, whatever) if that also works.

If you make changes, update `README.md` accordingly.
