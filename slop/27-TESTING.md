# Testing

I have linked a group on WhatsApp to a group on Signal. Then I tried to reply to group post on WhatsApp. This failed.

Here is the log:

```
2026/06/18 13:19:38 [DEBUG] WhatsApp Event Received on Channel:
{
  "Info": {
    "Chat": "31621222540-1555838023@g.us",
    "Sender": "281380638478587:42@lid",
    "IsFromMe": true,
    "IsGroup": true,
    "AddressingMode": "lid",
    "SenderAlt": "",
    "RecipientAlt": "",
    "BroadcastListOwner": "",
    "BroadcastRecipients": null,
    "ID": "3EB09D0404792AC2DD8499",
    "ServerID": 0,
    "Type": "text",
    "PushName": "Karel",
    "Timestamp": "2026-06-18T13:19:38+02:00",
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
      "text": "Hoewel ik AI vraag om de code te kloppen, is het design nog altijd stevig onder mijn controle 😉 Bovendien, ik kijk de AI-gegenereerde code altijd na. En Meta-AI gebruik ik niet, daar heb ik geen abonnement voor... daarentegen, bij Google heb ik geen limiet. Super powers!",
      "contextInfo": {
        "stanzaID": "3AF8FC9B1B2CEF4E750E",
        "participant": "145840010215630@lid",
        "quotedMessage": {
          "conversation": "Als het tenminste compatibel is met de AI waar het mee gemaakt is. Bij voorbaat geen Meta AI, anders kijkt Zuckerberg straks mee in onze Signal groep",
          "messageContextInfo": {}
        },
        "disappearingMode": {
          "initiator": 0,
          "trigger": 1
        }
      },
      "inviteLinkGroupTypeV2": 0
    },
    "messageContextInfo": {
      "messageSecret": "BhU83NfO4bz8AH3EwXZav3BPQd5yU9WLEvkkY9gw5pI=",
      "limitSharingV2": {
        "sharingLimited": true,
        "trigger": 1,
        "limitSharingSettingTimestamp": 1781256010326,
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
      "text": "Hoewel ik AI vraag om de code te kloppen, is het design nog altijd stevig onder mijn controle 😉 Bovendien, ik kijk de AI-gegenereerde code altijd na. En Meta-AI gebruik ik niet, daar heb ik geen abonnement voor... daarentegen, bij Google heb ik geen limiet. Super powers!",
      "contextInfo": {
        "stanzaID": "3AF8FC9B1B2CEF4E750E",
        "participant": "145840010215630@lid",
        "quotedMessage": {
          "conversation": "Als het tenminste compatibel is met de AI waar het mee gemaakt is. Bij voorbaat geen Meta AI, anders kijkt Zuckerberg straks mee in onze Signal groep",
          "messageContextInfo": {}
        },
        "disappearingMode": {
          "initiator": 0,
          "trigger": 1
        }
      },
      "inviteLinkGroupTypeV2": 0
    },
    "messageContextInfo": {
      "messageSecret": "BhU83NfO4bz8AH3EwXZav3BPQd5yU9WLEvkkY9gw5pI=",
      "limitSharingV2": {
        "sharingLimited": true,
        "trigger": 1,
        "limitSharingSettingTimestamp": 1781256010326,
        "initiatedByMe": false
      }
    }
  }
}
2026/06/18 13:19:38 [DEBUG] Raw msg.Message struct: extendedTextMessage:{text:"Hoewel ik AI vraag om de code te kloppen, is het design nog altijd stevig onder mijn controle 😉 Bovendien, ik kijk de AI-gegenereerde code altijd na. En Meta-AI gebruik ik niet, daar heb ik geen abonnement voor... daarentegen, bij Google heb ik geen limiet. Super powers!"  contextInfo:{stanzaID:"3AF8FC9B1B2CEF4E750E"  participant:"145840010215630@lid"  quotedMessage:{conversation:"Als het tenminste compatibel is met de AI waar het mee gemaakt is. Bij voorbaat geen Meta AI, anders kijkt Zuckerberg straks mee in onze Signal groep"  messageContextInfo:{}}  disappearingMode:{initiator:CHANGED_IN_CHAT  trigger:CHAT_SETTING}}  inviteLinkGroupTypeV2:DEFAULT}  messageContextInfo:{messageSecret:"\x06\x15<\xdc\xd7\xce\xe1\xbc\xfc\x00}\xc4\xc1vZ\xbfpOA\xderSՋ\x12\xf9$c\xd80\xe6\x92"  limitSharingV2:{sharingLimited:true  trigger:CHAT_SETTING  limitSharingSettingTimestamp:1781256010326  initiatedByMe:false}}
2026/06/18 13:19:38 [Sync WhatsApp -> Signal] Received message JID: 31621222540-1555838023@g.us, Sender: 281380638478587:42@lid, ID: 3EB09D0404792AC2DD8499
2026/06/18 13:19:38 [DEBUG] handleWhatsAppMessage: msg.Message=extendedTextMessage:{text:"Hoewel ik AI vraag om de code te kloppen, is het design nog altijd stevig onder mijn controle 😉 Bovendien, ik kijk de AI-gegenereerde code altijd na. En Meta-AI gebruik ik niet, daar heb ik geen abonnement voor... daarentegen, bij Google heb ik geen limiet. Super powers!"  contextInfo:{stanzaID:"3AF8FC9B1B2CEF4E750E"  participant:"145840010215630@lid"  quotedMessage:{conversation:"Als het tenminste compatibel is met de AI waar het mee gemaakt is. Bij voorbaat geen Meta AI, anders kijkt Zuckerberg straks mee in onze Signal groep"  messageContextInfo:{}}  disappearingMode:{initiator:CHANGED_IN_CHAT  trigger:CHAT_SETTING}}  inviteLinkGroupTypeV2:DEFAULT}  messageContextInfo:{messageSecret:"\x06\x15<\xdc\xd7\xce\xe1\xbc\xfc\x00}\xc4\xc1vZ\xbfpOA\xderSՋ\x12\xf9$c\xd80\xe6\x92"  limitSharingV2:{sharingLimited:true  trigger:CHAT_SETTING  limitSharingSettingTimestamp:1781256010326  initiatedByMe:false}}
2026/06/18 13:19:38 [DEBUG] Condition checks: Conversation=false (GetConversation=""), ExtendedTextMessage=true (GetText="Hoewel ik AI vraag om de code te kloppen, is het design nog altijd stevig onder mijn controle 😉 Bovendien, ik kijk de AI-gegenereerde code altijd na. En Meta-AI gebruik ik niet, daar heb ik geen abonnement voor... daarentegen, bij Google heb ik geen limiet. Super powers!"), AudioMessage=false, SecretEncryptedMessage=false
2026/06/18 13:19:38 [DEBUG] Forwarding to Signal: recipient=, group=aO/uf3Uk+H10SHqNE0iQZqSSQrqASOJHadr2iX/o+Pw=, message="[WhatsApp Group: 31621222540-1555838023@g.us] Karel: Hoewel ik AI vraag om de code te kloppen, is het design nog altijd stevig onder mijn controle 😉 Bovendien, ik kijk de AI-gegenereerde code altijd na. En Meta-AI gebruik ik niet, daar heb ik geen abonnement voor... daarentegen, bij Google heb ik geen limiet. Super powers!", attachments=[]
2026/06/18 13:19:38 Failed to forward WhatsApp message to Signal: signal-cli error -1: No recipients given
```

Apparently the group on Signal where the message should go, is not set. Fix it and check whether the same error might apply the other way around (Signal to WhatsApp).
