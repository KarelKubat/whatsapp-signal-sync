# Testing

I received a group message on WhatsApp. This is for a not linked group, so I expect the result to appear in my personal Signal `Note to Self`. But here I only got:

```
[WhatsApp Group: 31621222540-1555838023@g.us] Annemieke Meiborg: [WhatsApp received a message type that cannot be forwarded. Check your personal WhatsApp.]
```

The syncer saw the event and logged (because of `-debug`):

```
2026/06/17 16:09:23 [DEBUG] WhatsApp Event Received on Channel:
{
  "Info": {
    "Chat": "31621222540-1555838023@g.us",
    "Sender": "226581117186119@lid",
    "IsFromMe": false,
    "IsGroup": true,
    "AddressingMode": "lid",
    "SenderAlt": "41792976806@s.whatsapp.net",
    "RecipientAlt": "",
    "BroadcastListOwner": "",
    "BroadcastRecipients": null,
    "ID": "ACA2C40EB5F7296B0AF8CDDECFC92503",
    "ServerID": 0,
    "Type": "text",
    "PushName": "Annemieke Meiborg",
    "Timestamp": "2026-06-17T16:09:23+02:00",
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
    "extendedTextMessage": {
      "text": "@281380638478587 op verzoek van jou een berichtje zodat je je nieuwe software kunt testen 🤓",
      "previewType": 0,
      "contextInfo": {
        "mentionedJID": [
          "281380638478587@lid"
        ]
      },
      "inviteLinkGroupTypeV2": 0
    },
    "messageContextInfo": {
      "messageSecret": "rGvaWon4QppmLthpyyDLeeqodXTaIIR7x7rwmEJl+lw=",
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
    "extendedTextMessage": {
      "text": "@281380638478587 op verzoek van jou een berichtje zodat je je nieuwe software kunt testen 🤓",
      "previewType": 0,
      "contextInfo": {
        "mentionedJID": [
          "281380638478587@lid"
        ]
      },
      "inviteLinkGroupTypeV2": 0
    },
    "messageContextInfo": {
      "messageSecret": "rGvaWon4QppmLthpyyDLeeqodXTaIIR7x7rwmEJl+lw=",
      "limitSharingV2": {
        "sharingLimited": true,
        "trigger": 1,
        "limitSharingSettingTimestamp": 1781256010119,
        "initiatedByMe": false
      }
    }
  }
}
2026/06/17 16:09:23 [DEBUG] Raw msg.Message struct: extendedTextMessage:{text:"@281380638478587 op verzoek van jou een berichtje zodat je je nieuwe software kunt testen 🤓" previewType:NONE contextInfo:{mentionedJID:"281380638478587@lid"} inviteLinkGroupTypeV2:DEFAULT} messageContextInfo:{messageSecret:"\xack\xdaZ\x89\xf8B\x9af.\xd8i\xcb \xcby\xea\xa8ut\xda \x84{Ǻ\xf0\x98Be\xfa\\" limitSharingV2:{sharingLimited:true trigger:CHAT_SETTING limitSharingSettingTimestamp:1781256010119 initiatedByMe:false}}
```

Why was the message not forwarded in a better way so I could read it on Signal?

Investigate this, and fix it.
