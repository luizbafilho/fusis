---
title: POST destination
permalink: /post-destination
layout: default
---

# POST destination

## /services/:service_name/destinations

Add a destination to a service

## Options

Name | Type | Description
:--- | :--- | :---
name | string | Destination name
port | number | Destination port
host | ip | Destination ip address
weight | number | Destination weight

## Sample response

**HTTP/1.1 200 OK**

```json
{
  "Host": "10.0.0.6",
  "Mode": "masq",
  "Name": "test2",
  "Port": 80,
  "ServiceId": "filmes2",
  "Weight": 1
}
```
