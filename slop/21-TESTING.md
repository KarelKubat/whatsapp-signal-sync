# Testing

I got a Signal chat with an attachment and forwarding to WhatsApp failed. On WhatsApp I see only: 

```
[Signal attachment forward failed: open : no such file or directory]
```

The log states:

```
026/06/17 15:34:54 [DEBUG] Signal Event Received on Socket:
{
  "params": {
    "envelope": {
      "source": "+41792976806",
      "sourceName": "Annemieke Meiborg",
      "sourceNumber": "+41792976806",
      "sourceUuid": "37714a2c-348c-4802-b60a-f44a42366cbf",
      "timestamp": 1781703293985,
      "dataMessage": {
        "timestamp": 1781703293985,
        "message": "",
        "quote": null,
        "groupInfo": null,
        "attachments": [
          {
            "contentType": "image/jpeg",
            "filename": "",
            "id": "J1Lg9QX34bqsc8Q8xezD.jpg",
            "size": 302486,
            "storedFilename": ""
          }
        ]
      },
      "syncMessage": null
    },
    "account": "+41798766719"
  }
}
2026/06/17 15:34:54 [Sync Signal -> WhatsApp] Received message from Source: +41792976806, Msg:
2026/06/17 15:34:54 Failed to read Signal attachment: open : no such file or directory
```

Investigate and fix it.
