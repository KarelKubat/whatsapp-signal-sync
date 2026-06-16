# whatsapp-signal-sync

## Purpose

I want you to design for me a CLI program that, once started, syncs my conversations between WhatsApp and Signal. The requirements are:

- Startup authentication: 

  - The program will have to authenticate with WhatsApp and Signal using my credentials. I know that for WhatsApp it is possible to render a QR code, to be scanned by a mobile WhatsApp application, to authenticate a new device. This is acceptable for WhatsApp but you must find something similar to be used for Signal.
  - Instead of rendering a QR code and asking me to authenticate a new device, it is also acceptable that a first run asks to generate an API key or an access key, and to store that key for the next runs in a local configuration file (JSON or YAML formats are acceptable), or in environment variables, e.g. `WHATSAPP_API_KEY=some-key-string`.
  - If possible, allow for both: either with user intervention to scan QR codes and to confirm something, **and** a local configuration file or environment variables.

- Startup group selection:

  - On each channel (WhatApp or Signal) the program must be able to list the groups that I am member of. It must present the groups and ask how to link them, so that if a group message comes in on one channel, it can be forwarded to a group on the other channel. 
  - This setting must be stored in the local configuration file.

- Functional requirements

  - For simple text messages, only sent to me: When a message comes in on one of the channels (WhatsApp or Signal) then the program must forward this to the other channel, with a clear text at the top stating that this was forwarded. Example of a forward to from WhatsApp to Signal: *WhatsApp received the personal message: original message text*
  
  - For group text messages, sent to a group that I am member of: 
    - If there is a corresponding message on the other channel, then the program must forward the message to that other channel. The forwarded message must state this, example: *WhatApp received the group message: orginal message text*
    - If the group that that received the message is not linked to a group on the other channel, then the message must be forwarded to my personal account (non-group) on that other channel, clearly stating so in the forward. Example: *WhatsApp received the group message, but there is no corresponding group on Signal: original message text*

  - Images and video: The program must be able to forward images and video similar to text messages.

  - For other messages, like polls: If you can find a way to also forward these, then do so. If not, send a text message to the other channel stating so. Example: *WhatsApp received a message that cannot be forwarded. Check your personal WhatsApp.*

## How to do it

Before crafting Go code, create a file `DESIGN-it-0.md` where you outline how to do it. When that design is ready, I will review it and we will go from there.
