# Testing

I received a group message on Signal. On screen it shows up as:

```
Sender: Message
```

This got forwarded to my personal WhatsApp. But there it shows as:

```
[Signal Group: group-ID] Sender (in ) : Message
```

So the forwarding in itself works but the layout can be improved:

1. There is a random space between `(in )` and `:`. Can you remove it?
2. Can you put the WhatsApp group name behind the `in`. If the group name cannot be determined, leave out the entire `(in )` substring.
