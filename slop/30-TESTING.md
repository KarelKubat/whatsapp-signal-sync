# Testing

I got a message on WhatsApp that could not be delivered to the linked Signal group. It ended up on Signal with:

```
[WhatsApp received a message type that cannot be forwarded. Check there.]
```

Here is the log.

```
2026/06/18 15:04:28 [DEBUG] WhatsApp Event Received on Channel:
{
  "Info": {
    "Chat": "31621222540-1555838023@g.us",
    "Sender": "267310258479129@lid",
    "IsFromMe": false,
    "IsGroup": true,
    "AddressingMode": "lid",
    "SenderAlt": "31652435825@s.whatsapp.net",
    "RecipientAlt": "",
    "BroadcastListOwner": "",
    "BroadcastRecipients": null,
    "ID": "AC42A9729662F3323D786ACA232EA9B8",
    "ServerID": 0,
    "Type": "text",
    "PushName": "Willem",
    "Timestamp": "2026-06-18T15:04:26+02:00",
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
    "conversation": "Wie zorgt hier voor de begrijpelijke ondertiteling voor normale stervelingen?",
    "messageContextInfo": {
      "messageSecret": "NjC/Kx4bkJvDgyG6Bq8rB7NYKEygGFdk247cU0lSFn4=",
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
    "conversation": "Wie zorgt hier voor de begrijpelijke ondertiteling voor normale stervelingen?",
    "messageContextInfo": {
      "messageSecret": "NjC/Kx4bkJvDgyG6Bq8rB7NYKEygGFdk247cU0lSFn4=",
      "limitSharingV2": {
        "sharingLimited": true,
        "trigger": 1,
        "limitSharingSettingTimestamp": 1781256010119,
        "initiatedByMe": false
      }
    }
  }
}
2026/06/18 15:04:28 [DEBUG] Raw msg.Message struct: conversation:"Wie zorgt hier voor de begrijpelijke ondertiteling voor normale stervelingen?" messageContextInfo:{messageSecret:"60\xbf+\x1e\x1b\x90\x9bÃ!\xba\x06\xaf+\x07\xb3X(L\xa0\x18Wdێ\xdcSIR\x16~" limitSharingV2:{sharingLimited:true trigger:CHAT_SETTING limitSharingSettingTimestamp:1781256010119 initiatedByMe:false}}
2026/06/18 15:04:28 [Sync WhatsApp -> Signal] Discarding duplicate or older message (msgTime: 1781787866, lastTime: 1781787866, msgID: AC42A9729662F3323D786ACA232EA9B8, lastID: AC42A9729662F3323D786ACA232EA9B8)
```

Why did this happen? Fix it. My own remarks:
- I suspect that the message was dropped by the sync program because it incorrectly determined that it was a duplicate, which in not the case here.
- If the duplicate checking is brittle, maybe use a state database where you store IDs and/or timestamps of previously forwarded messages.
