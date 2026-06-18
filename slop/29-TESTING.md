# Testing

When I reply to a group message on WhatsApp, I see on Signal only the text of my reply. Can you make it so that also the message is shown to which I replied? This should be for linked groups, unlinked, and for both directions (WhatsApp to Signal and Signal to WhatsApp).

Also, remove any `[...]` headers from forwarding when such headers are only generated with `-debug`. If the headers are needed for routing, then leave them, but otherwise, make the messages as clean as possible.

I see that you are checking for `[` and `]` to see whether a message has a header. That is potentially brittle and I am not entirely happy with it. Here are some suggestions for improvement:
- If you are parsing messages for headers enclosed in `[` and `]`, then make that parsing as robust as possible. It must not happen that a user message that starts with `[` leads to a fault.
- If possible, embed routing information in another field than the message text itself. Maybe there is a field that is not displayed that could be used for it? I have not checked, analyze this option.
