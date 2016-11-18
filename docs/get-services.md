---
title: GET services
permalink: /get-services
layout: default
---

# GET services

## /services

List all services

## Sample response

**HTTP/1.1 200 OK**

```json
[
  {
    "Name": "billing-service",
    "Host": "192.168.0.1",
    "Port": 80,
    "Protocol": "tcp",
    "Scheduler": "rr",
    "Mode": "masq",
    "Persistent": 0,
    "Destinations": [
      {
        "Name": "billing-node1",
        "Host": "10.0.0.2",
        "Port": 8080,
        "Weight": 1,
        "Mode": "masq",
        "ServiceId": "filmes2"
      },
      {
        "Name": "billing-node2",
        "Host": "10.0.0.3",
        "Port": 8080,
        "Weight": 1,
        "Mode": "masq",
        "ServiceId": "filmes2"
      }
    ]
  }
]
```
