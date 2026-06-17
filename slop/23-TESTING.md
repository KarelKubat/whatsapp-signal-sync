# Testing

I got an audio message on WhatsApp. The debug log is:

```
2026/06/17 16:49:33 [DEBUG] WhatsApp Event Received on Channel:
{
  "Info": {
    "Chat": "15818465276038@lid",
    "Sender": "15818465276038@lid",
    "IsFromMe": false,
    "IsGroup": false,
    "AddressingMode": "",
    "SenderAlt": "31611280180@s.whatsapp.net",
    "RecipientAlt": "",
    "BroadcastListOwner": "",
    "BroadcastRecipients": null,
    "ID": "AC7F564BC0E162C1401E3F78AFC5DE15",
    "ServerID": 0,
    "Type": "text",
    "PushName": "Monique",
    "Timestamp": "2026-06-17T16:49:33+02:00",
    "Category": "",
    "Multicast": false,
    "MediaType": "",
    "Edit": "7",
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
    "protocolMessage": {
      "key": {
        "remoteJID": "281380638478587@lid",
        "fromMe": true,
        "ID": "AC02B0BA89C34014CA4C7B0838574C1D"
      },
      "type": 0
    },
    "messageContextInfo": {
      "deviceListMetadata": {
        "senderKeyHash": "FT3aEfz7JWIvaw==",
        "senderTimestamp": 1781415936,
        "recipientKeyHash": "vABF7YwT3yhbBQ==",
        "recipientTimestamp": 1781619504
      },
      "deviceListMetadataVersion": 2
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
    "protocolMessage": {
      "key": {
        "remoteJID": "281380638478587@lid",
        "fromMe": true,
        "ID": "AC02B0BA89C34014CA4C7B0838574C1D"
      },
      "type": 0
    },
    "messageContextInfo": {
      "deviceListMetadata": {
        "senderKeyHash": "FT3aEfz7JWIvaw==",
        "senderTimestamp": 1781415936,
        "recipientKeyHash": "vABF7YwT3yhbBQ==",
        "recipientTimestamp": 1781619504
      },
      "deviceListMetadataVersion": 2
    }
  }
}
2026/06/17 16:49:33 [DEBUG] Raw msg.Message struct: protocolMessage:{key:{remoteJID:"281380638478587@lid"  fromMe:true  ID:"AC02B0BA89C34014CA4C7B0838574C1D"}  type:REVOKE}  messageContextInfo:{deviceListMetadata:{senderKeyHash:"\x15=\xda\x11\xfc\xfb%b/k"  senderTimestamp:1781415936  recipientKeyHash:"\xbc\x00E\xed\x8c\x13\xdf([\x05"  recipientTimestamp:1781619504}  deviceListMetadataVersion:2}
2026/06/17 16:49:33 [Sync WhatsApp -> Signal] Received message JID: 15818465276038@lid, Sender: 15818465276038@lid, ID: AC7F564BC0E162C1401E3F78AFC5DE15
2026/06/17 16:49:33 [DEBUG] handleWhatsAppMessage: msg.Message=protocolMessage:{key:{remoteJID:"281380638478587@lid"  fromMe:true  ID:"AC02B0BA89C34014CA4C7B0838574C1D"}  type:REVOKE}  messageContextInfo:{deviceListMetadata:{senderKeyHash:"\x15=\xda\x11\xfc\xfb%b/k"  senderTimestamp:1781415936  recipientKeyHash:"\xbc\x00E\xed\x8c\x13\xdf([\x05"  recipientTimestamp:1781619504}  deviceListMetadataVersion:2}
2026/06/17 16:49:33 [DEBUG] Condition checks: Conversation=false (GetConversation=""), ExtendedTextMessage=false (GetText="")
2026/06/17 16:49:33 [DEBUG] Forwarding to Signal: recipient=+41798766719, group=, message="[WhatsApp Direct: 15818465276038@lid] Monique: [WhatsApp received a message type that cannot be forwarded. Check your personal WhatsApp.]", attachments=[]
2026/06/17 16:49:33 Forwarded WhatsApp message to Signal successfully.
```

Despite the fact that the sender later deleted their WhatsApp message, the audio never got to my Signal. Instead I saw on Signal that the message type could not be forwarded.

Can you implement audio messaging?
