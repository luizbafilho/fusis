---
title: POST service
permalink: /post-service
layout: default
---

# POST service

## /service

Create a new service

## Options

Name | Type | Description
:--- | :--- | :---
name | string | Service name
port | number | Service port
mode | string | Service [forwarding mode](forwarding-modes)
protocol | string | Service protocol
scheduler | string | Service [scheduler](scheduling-modes)
persistent | number | Persistence timeout in seconds

## Sample response

**HTTP/1.1 200 OK**

```json
{
  "Name": "filmes4",
  "Host": "192.168.0.4",
  "Port": 80,
  "Protocol": "tcp",
  "Scheduler": "rr",
  "Mode": "nat",
  "Persistent": 0
}
```
