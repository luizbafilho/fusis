---
title: Forwarding Modes
permalink: /forwarding-modes
layout: default
---

# Forwarding Modes

There are three forwarding methods available. These are used to determine how incoming requests to cluster nodes will be relayed.

* Network Address Translation (nat)
* Direct Routing (route)
* IP Tunneling (tunnel)

## Network Address Translation

Use the Linux Kernel ability to translate IP Address and ports as packages them pass through the Kernel.

Receive a request to a service and forwards this request to a real server. The real server then replies the request by sending the packet back to Fusis, that in turn, translates (NAT) the real server address back to the service address. This masks all the real servers addresses and ports behind a unique address and port bound to Fusis.

![NAT](images/nat.png)

## Direct Routing

In this mode, forwards all incoming requests to the real servers and their replies go straight back to the client - without passing back through Fusis.

![DSR](images/dsr.png)

## IP Tunneling

This mode works like Direct Routing but it encapsulates the incoming packets into another IP packet. The original packet becomes the payload of the newly generated packet. Making possible to forward​ the packets to any other network because there would be no need to be on the same CIDR.
