# Testing

The syncer log shows the following message which is a reply to a group chat on WhatsApp. This got correctly sent to my Signal `Note to Self`:

```
2026/06/17 22:28:50 [DEBUG] WhatsApp Event Received on Channel:
{
  "Info": {
    "Chat": "31621222540-1555838023@g.us",
    "Sender": "266889401995473@lid",
    "IsFromMe": false,
    "IsGroup": true,
    "AddressingMode": "lid",
    "SenderAlt": "31613066565@s.whatsapp.net",
    "RecipientAlt": "",
    "BroadcastListOwner": "",
    "BroadcastRecipients": null,
    "ID": "3A2003A1CFF471A991F7",
    "ServerID": 0,
    "Type": "text",
    "PushName": "Martinus Meiborg",
    "Timestamp": "2026-06-17T22:28:50+02:00",
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
    "conversation": "Naar dan kan je weer aan een AI vragen om die code uit te leggen.",
    "messageContextInfo": {
      "messageSecret": "VPr0OKP9QOCTbclFu0d2ZOx7cNHW08oDng9rSngM+ns=",
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
    "conversation": "Naar dan kan je weer aan een AI vragen om die code uit te leggen.",
    "messageContextInfo": {
      "messageSecret": "VPr0OKP9QOCTbclFu0d2ZOx7cNHW08oDng9rSngM+ns=",
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

Then the sender on WhatsApp edited their message, changing the text `Naar dan kan je..` into `Maar dan kan je..`. The log states:

```
2026/06/17 22:29:17 [DEBUG] WhatsApp Event Received on Channel:
{
  "Info": {
    "Chat": "31621222540-1555838023@g.us",
    "Sender": "266889401995473@lid",
    "IsFromMe": false,
    "IsGroup": true,
    "AddressingMode": "lid",
    "SenderAlt": "31613066565@s.whatsapp.net",
    "RecipientAlt": "",
    "BroadcastListOwner": "",
    "BroadcastRecipients": null,
    "ID": "3A5940D0F4B5C3B44F2B",
    "ServerID": 0,
    "Type": "text",
    "PushName": "Martinus Meiborg",
    "Timestamp": "2026-06-17T22:29:17+02:00",
    "Category": "",
    "Multicast": false,
    "MediaType": "",
    "Edit": "1",
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
    "messageContextInfo": {
      "limitSharingV2": {
        "sharingLimited": true,
        "trigger": 1,
        "limitSharingSettingTimestamp": 1781256010119,
        "initiatedByMe": false
      }
    },
    "secretEncryptedMessage": {
      "targetMessageKey": {
        "remoteJID": "31621222540-1555838023@g.us",
        "fromMe": true,
        "ID": "3A2003A1CFF471A991F7"
      },
      "encPayload": "nTd8xYtL+Crqx/Tvhyw24xQlC4iOTx1A39kopeBQNPvR1s9WKfE/Qad4j0hg/rjVdpaCV02t+kIfu13exVtrC4krHBiKaZKXeEFNr/RbVbEOO8BC4dCnCOlM9SLUmOmFkPjT2pDZGM/GBgBYI3EJd/g1jGNhGFnPjxo9cFklskosdFSWKOApJaFF38vPpwABDg2Rbj9wehXVx9ltl/xFGgfDVcF0kDeAm99IP18m9RZbX/o8GCazaFQ4lS4cQ34LUk58CdsU2mQdMjH3",
      "encIV": "/xU8ewn0hFs2rRoH",
      "secretEncType": 2
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
    "messageContextInfo": {
      "limitSharingV2": {
        "sharingLimited": true,
        "trigger": 1,
        "limitSharingSettingTimestamp": 1781256010119,
        "initiatedByMe": false
      }
    },
    "secretEncryptedMessage": {
      "targetMessageKey": {
        "remoteJID": "31621222540-1555838023@g.us",
        "fromMe": true,
        "ID": "3A2003A1CFF471A991F7"
      },
      "encPayload": "nTd8xYtL+Crqx/Tvhyw24xQlC4iOTx1A39kopeBQNPvR1s9WKfE/Qad4j0hg/rjVdpaCV02t+kIfu13exVtrC4krHBiKaZKXeEFNr/RbVbEOO8BC4dCnCOlM9SLUmOmFkPjT2pDZGM/GBgBYI3EJd/g1jGNhGFnPjxo9cFklskosdFSWKOApJaFF38vPpwABDg2Rbj9wehXVx9ltl/xFGgfDVcF0kDeAm99IP18m9RZbX/o8GCazaFQ4lS4cQ34LUk58CdsU2mQdMjH3",
      "encIV": "/xU8ewn0hFs2rRoH",
      "secretEncType": 2
    }
  }
}
2026/06/17 22:29:17 [DEBUG] Raw msg.Message struct: messageContextInfo:{limitSharingV2:{sharingLimited:true  trigger:CHAT_SETTING  limitSharingSettingTimestamp:1781256010119  initiatedByMe:false}}  secretEncryptedMessage:{targetMessageKey:{remoteJID:"31621222540-1555838023@g.us"  fromMe:true  ID:"3A2003A1CFF471A991F7"}  encPayload:"\x9d7|ŋK\xf8*\xea\xc7\xf4\xef\x87,6\xe3\x14%\x0b\x88\x8eO\x1d@\xdf\xd9(\xa5\xe0P4\xfb\xd1\xd6\xcfV)\xf1?A\xa7x\x8fH`\xfe\xb8\xd5v\x96\x82WM\xad\xfaB\x1f\xbb]\xde\xc5[k\x0b\x89+\x1c\x18\x8ai\x92\x97xAM\xaf\xf4[U\xb1\x0e;\xc0B\xe1Ч\x08\xe9L\xf5\"Ԙ酐\xf8\xd3ڐ\xd9\x18\xcf\xc6\x06\x00X#q\tw\xf85\x8cca\x18YϏ\x1a=pY%\xb2J,tT\x96(\xe0)%\xa1E\xdf\xcbϧ\x00\x01\x0e\r\x91n?pz\x15\xd5\xc7\xd9m\x97\xfcE\x1a\x07\xc3U\xc1t\x907\x80\x9b\xdfH?_&\xf5\x16[_\xfa<\x18&\xb3hT8\x95.\x1cC~\x0bRN|\t\xdb\x14\xdad\x1d21\xf7"  encIV:"\xff\x15<{\t\xf4\x84[6\xad\x1a\x07"  secretEncType:MESSAGE_EDIT}
2026/06/17 22:29:17 [Sync WhatsApp -> Signal] Received message JID: 31621222540-1555838023@g.us, Sender: 266889401995473@lid, ID: 3A5940D0F4B5C3B44F2B
2026/06/17 22:29:17 [DEBUG] handleWhatsAppMessage: msg.Message=messageContextInfo:{limitSharingV2:{sharingLimited:true  trigger:CHAT_SETTING  limitSharingSettingTimestamp:1781256010119  initiatedByMe:false}}  secretEncryptedMessage:{targetMessageKey:{remoteJID:"31621222540-1555838023@g.us"  fromMe:true  ID:"3A2003A1CFF471A991F7"}  encPayload:"\x9d7|ŋK\xf8*\xea\xc7\xf4\xef\x87,6\xe3\x14%\x0b\x88\x8eO\x1d@\xdf\xd9(\xa5\xe0P4\xfb\xd1\xd6\xcfV)\xf1?A\xa7x\x8fH`\xfe\xb8\xd5v\x96\x82WM\xad\xfaB\x1f\xbb]\xde\xc5[k\x0b\x89+\x1c\x18\x8ai\x92\x97xAM\xaf\xf4[U\xb1\x0e;\xc0B\xe1Ч\x08\xe9L\xf5\"Ԙ酐\xf8\xd3ڐ\xd9\x18\xcf\xc6\x06\x00X#q\tw\xf85\x8cca\x18YϏ\x1a=pY%\xb2J,tT\x96(\xe0)%\xa1E\xdf\xcbϧ\x00\x01\x0e\r\x91n?pz\x15\xd5\xc7\xd9m\x97\xfcE\x1a\x07\xc3U\xc1t\x907\x80\x9b\xdfH?_&\xf5\x16[_\xfa<\x18&\xb3hT8\x95.\x1cC~\x0bRN|\t\xdb\x14\xdad\x1d21\xf7"  encIV:"\xff\x15<{\t\xf4\x84[6\xad\x1a\x07"  secretEncType:MESSAGE_EDIT}
2026/06/17 22:29:17 [DEBUG] Condition checks: Conversation=false (GetConversation=""), ExtendedTextMessage=false (GetText=""), AudioMessage=false
2026/06/17 22:29:17 [DEBUG] Forwarding to Signal: recipient=+41798766719, group=, message="[WhatsApp Group: 31621222540-1555838023@g.us] Martinus Meiborg: [WhatsApp received a message type that cannot be forwarded. Check your personal WhatsApp.]", attachments=[]
```

This event got delivered to my Signal as a message that the robot could not forward. Can you fix that?  If not, a better message would be that "<user> edited their message on WhatsApp, check there" or something similar, so that it has more context than the current text which basically only states that it's not deliverable.
