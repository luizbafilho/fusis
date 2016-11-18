---
title: Cluster Modes
permalink: /cluster-modes
layout: default
---

# Cluster Modes

Fusis can be deployed in two different cluster modes (anycast or unicast), which one to use, it depends on your requirements and network capabilities.

* `ANYCAST` - Provides full support to [Anycast](https://en.wikipedia.org/wiki/Anycast) VIPs, once configured advertises the [VIPs](https://en.wikipedia.org/wiki/Virtual_IP_address) to the network using BGP. To that happen correctly you have to make sure that your BGP Peers accepts routes from hosts address you deployed Fusis. There no need to external tools like `quagga` or `bird`, Fusis daemon advertises the VIPs without any dependency. All nodes balance​ traffic to the real servers.

* `UNICAST` - When you do not need or can't use Anycast you should use the `Unicast` mode. In this mode only one of the nodes balances the traffic at the time, the others keep running as backup, once the leader node fails one of the other nodes assumes the leadership of the cluster and then start balances the traffic.
