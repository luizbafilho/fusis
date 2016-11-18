---
title: Configuration
permalink: /configuration
layout: default
---

# Configuration

Fusis have several options that can be set through a configuration file. This page will describe all of them in detail.

## Options

### name
Defines the unique name that identifies the node in the cluster

### cluster-mode
Defines the cluster mode used. Fusis can be initialized in two modes `unicast` and `anycast` to further details on what they are and when to use them, please visit the [Cluster Modes](cluster-modes) page.

### join
Defines the IPs of the other members of the cluster so that the node can join the cluster upon initialization.

### data-path
Defines the path used to store information on the nodes and Raft logs and snapshots.

## **[interfaces]**
This section defines the network interfaces used by Fusis. For better performance, we recommend the use of two network interfaces, usually one has public access, and the other has internal access. There is no problem to run Fusis with only one network interface, but it can become a traffic bottleneck depending on the network load.

### inbound
Defines the inbound interface which is the one that receives the client requests. Usually, this one is connected to a public network. If not set the default value is `eth0`.

### outbound
Defines the outbound interface which is the one that forward the received requests. Usually, this one is connected to a private network. If not set the default value is `eth0`.

## **[ipam]**
This section defines the IP ranges used by Fusis to allocate new Virtual IPs when a service is created. The ranges must be reserved, and no one should be using them but Fusis.

### ranges
Defines the IP ranges used by fusis. Here you can defines as many ranges you want, just make sure they don't overlap.

## **[ports]**
This section defines the network ports used for different services used by Fusis.

### api
Defines the port used by the HTTP API server. Defaults to 8000

### raft
Defines the port used by the consensus protocol. Defaults to 4382

### serf
Defines the port used by the gossip protocol. Defaults to 7940

## **[metrics]**
This section defines the metrics configuration.

### publisher
The defines the publisher used by the metrics collector to push the data out. For now, it only supports pushing metrics to `Logstash` but support to `Influxdb` and `Statsd` is planned.

### [metrics.params]
Defines the `host` and `port` used to push metrics out.

### [metrics.extras]
This option is used if you want to append extra information to the metric sent. Here you can add extra information like datacenter or node name for example.

## **[bgp]**
This section defines `BGP` configuration required when running the cluster in  `anycast` mode.

### as
Defines the 'autonomous system number'.

### router-id
Defines the router id.

### [[bgp.neighbors]]
Defines an array of neighbors that can be configured to communicate to Fusis.

### address
Defines the address of the neighbor router

### peer-as
Defines the 'autonomous system number' of the neighbor router

## Sample file

This an​ example of a complete config file setting all options available.

```toml
# Sample Fusis configuration file

name = "example"
cluster-mode = "anycast"
join = ["10.0.0.2", "186.23.22.1"]
data-path = "/etc/fusis"

[interfaces]
inbound = "eth0"
outbound = "eth1"

[ipam]
ranges = ["192.168.0.0/28"]

[ports]
raft = 4382
serf = 7940
api = 8000

[metrics]
publisher = "logstash"

  [metrics.params]
  host = "0.0.0.0"
  port = 8989

  [metrics.extras]
  client = "fusis"

[bgp]
as = 100
router-id = "192.168.151.182"

  [[bgp.neighbors]]
  address = "192.168.151.178"
  peer-as = 100
```
