Fusis Balancer
======

Fusis Balancer is a dynamic Layer 4 Load Balancer powered by [IPVS](http://www.linuxvirtualserver.org/) and [Serf](https://www.serfdom.io/).

Running the Fusis Agent in your servers, lets the balancer detect new nodes and add new routes to them.

There is also a HTTP API to manages your services dynamically.

### IPVS
IPVS (IP Virtual Server) implements transport-layer load balancing directly in the Linux Kernel. It's being aroung since 1999 and its very stable and battle tested. Used by many companies like Google, Facebook, Github, Soundcloud and many others.

IPVS it is a amazing piece of software that few people really know of it, and that is probably because its not that easy to use. It requires some knowledge about networking and its necessary to do some network configuration in order to make it work correctly.

### Serf
Serf is solution for distributed cluster management, message delivery and failure detection. It is one of the bases of [Consul](https://www.consul.io/).


## Why?
The whole goal of this project it is to bring a accessible way to use IPVS.

Fusis Balancer will be responsible for detecting new/failed nodes and add/remove routes to it. It will automatically configure the network in order to make everything work.

## State
This project it's under heavy development, it's not usable yet, but you can **Star** :star: the project and follow the updates.

# Installation

There is compilation and runtime dependency on [libnl](https://www.infradead.org/~tgr/libnl/).
On a Debian/Ubuntu style system, you should be able to prepare for building by running:

``
apt-get install libnl-3-dev libnl-genl-3-dev
``

Get this project into GOPATH:

```
go get -v github.com/luizbafilho/fusis
```

Also, get these projects:

```
go get -v github.com/hashicorp/errwrap
go get -v github.com/hashicorp/go-multierror
go get -v github.com/miekg/dns
```

Install `govendor` and get the dependencies.

```
go get github.com/kardianos/govendor
govendor add +e
```

You will need at least **Go 1.5**.


## Installing IPVS

IPVS is a Kernel module. Install it using modprobe to enable it on the kernel:

```
# Enables ipvs on kernel
$> modprobe ip_vs

# Also install the IPVS command line tool
$> sudo apt-get install ipvsadm
```

To use IPVS, you must enable the IP forwarding parameter for kernel
```
$> sudo sysctl -w net.ipv4.ip_forward=1
```

## Running the project


