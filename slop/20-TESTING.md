# Testing

I received a group message from WhatsApp which is just one line of plain text. This is for a not linked group, so I expect the message to appear in my personal Signal `Note to Self` but it never made it there. Instead I got on my Signal: 

```
[WhatsApp Group: 31621222540-1555838023@g.us] Gerrie Meiborg [WhatsApp received a message type that cannot be forwarded. Check your personal WhatsApp.]
```

The syncer saw the event and logged (because of `-debug`):

```
2026/06/17 14:07:15 [DEBUG] WhatsApp Event Received on Channel:
{
  "Info": {
    "Chat": "31621222540-1555838023@g.us",
    "Sender": "23948821528822@lid",
    "IsFromMe": false,
    "IsGroup": true,
    "AddressingMode": "lid",
    "SenderAlt": "31610352329@s.whatsapp.net",
    "RecipientAlt": "",
    "BroadcastListOwner": "",
    "BroadcastRecipients": null,
    "ID": "AC9623C2625FD56AB8C46A86C9FAEE62",
    "ServerID": 0,
    "Type": "text",
    "PushName": "Gerrie Meiborg",
    "Timestamp": "2026-06-17T14:07:14+02:00",
    "Category": "",
    "Multicast": false,
    "MediaType": "",
    "Edit": "",
    "MsgBotInfo": {
      "EditType": "",
      "EditTargetID": "",
      "EditSenderTimestampMS": "0001-01-01T00:00:00Z"
    },
    "MsgMetaInfo": {
      "TargetID": "",
      "TargetSender": "",
      "TargetChat": "",
      "DeprecatedLIDSession": null,
      "ThreadMessageID": "",
      "ThreadMessageSenderJID": ""
    },
    "VerifiedName": null,
    "DeviceSentMeta": null
  },
  "Message": {
    "conversation": "Ben heel benieuwd hoe zo'n virtuele bal naar binnen glijdt!",
    "messageContextInfo": {
      "messageSecret": "fsNZl3McmfmVnQGZGqM/JpO57IMW56kxNN4VNSUXSJ0=",
      "limitSharingV2": {
        "sharingLimited": true,
        "trigger": 1,
        "limitSharingSettingTimestamp": 1781256010119,
        "initiatedByMe": false
      }
    }
  },
  "IsEphemeral": false,
  "IsViewOnce": false,
  "IsViewOnceV2": false,
  "IsViewOnceV2Extension": false,
  "IsDocumentWithCaption": false,
  "IsLottieSticker": false,
  "IsBotInvoke": false,
  "IsEdit": false,
  "SourceWebMsg": null,
  "UnavailableRequestID": "",
  "RetryCount": 0,
  "NewsletterMeta": null,
  "RawMessage": {
    "conversation": "Ben heel benieuwd hoe zo'n virtuele bal naar binnen glijdt!",
    "messageContextInfo": {
      "messageSecret": "fsNZl3McmfmVnQGZGqM/JpO57IMW56kxNN4VNSUXSJ0=",
      "limitSharingV2": {
        "sharingLimited": true,
        "trigger": 1,
        "limitSharingSettingTimestamp": 1781256010119,
        "initiatedByMe": false
      }
    }
  }
}
```

Why was the message which is just text not get forwarded as just text?

Investigate this, and fix it.
