# Headers

I see that you are checking for `[` and `]` to see whether a message has a header. That is potentially brittle and I am not entirely happy with it. Here are some suggestions for improvement:
- If you are parsing messages for headers enclosed in `[` and `]`, then make that parsing as robust as possible. It must not happen that a user message that starts with `[` leads to a fault.
- If possible, embed routing information in another field than the message text itself. Maybe there is a field that is not displayed that could be used for it? I have not checked, analyze this option.
