Fusis Balancer
======

Fusis Balancer is a dynamic Layer 4 Load Balancer powered by [IPVS](http://www.linuxvirtualserver.org/) and [Serf](https://www.serfdom.io/).

Running the Fusis Agent in your servers, lets the load balancer detect new nodes and route traffic to them.

There is also a HTTP API to manage your services dynamically.

### IPVS
IPVS (IP Virtual Server) implements transport-layer load balancing directly in the Linux Kernel. It has been around since 1999 and is very stable/battle tested. Used by many companies such as Google, Facebook, Github, Soundcloud and so on.

IPVS is a amazing piece of software that few people really got to know, probably because its not that easy to use as it requires some some network configuration and a deep knowledge of it in order to make it work correctly.

### Serf
Serf is solution for distributed cluster management, message delivery and failure detection. It's one of the bases of [Consul](https://www.consul.io/).

## Why?
The whole goal of this project is to provide an easy way to use IPVS.

Fusis Load Balancer will be responsible for detecting new/failed nodes and add/remove routes to them. It will automatically configure the network to do so.

## State
This project it's under heavy development, it's not usable yet, but you can **Star** :star: the project and follow the updates.

# Installation

There is compilation and runtime dependency on [libnl](https://www.infradead.org/~tgr/libnl/).
On a Debian based system, you should be able to build it by running:

``` bash
sudo apt-get install libnl-3-dev libnl-genl-3-dev
```

Get this project into GOPATH:

``` bash
go get -v github.com/luizbafilho/fusis
```

And it's dependencies:

``` bash
make restore
```
You'll need **Go 1.5** or later;

## Installing IPVS

IPVS is a Kernel module. So, all you need to do is enable it via modprobe:
``` bash
sudo modprobe ip_vs
```

While you're at it, install the IPVS command line tool too:
``` bash
sudo apt-get install ipvsadm
```

And enable ipv4 forwarding with:
``` bash
sudo sysctl -w net.ipv4.ip_forward=1
```

## Running the project

Now that you have IPVS and fusis installed, run the project:

``` bash
# Remenber, fusis binary is at $GOPATH/bin/fusis. So, add it to your system PATH
sudo fusis balancer --iface eth0
```
You should see something like:
`[GIN-debug] Listening and serving HTTP on :8000`

From another host, send a HTTP request to the API querying for available services available:
``` bash
curl -i {IP OF FUSIS HOST}:8000/services
```
And you should get:
```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Thu, 07 Apr 2016 21:23:18 GMT
Content-Length: 3

[]
```

Just for testing purposes, lets add a route to a fake IPv4 by runnging this on the fusis host:

``` bash
sudo ipvsadm -A -t 10.0.0.1:80 -s rr
```

Then, make another request:

``` bash
curl -i {FUSIS_HOST_IPV4}:8000/services
```

You will get that same route you just created as a response:
```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Thu, 07 Apr 2016 22:08:42 GMT
Content-Length: 94

[{"Name":"","Host":"10.0.0.1","Port":80,"Protocol":"tcp","Scheduler":"rr","Destinations":[]}]
```
