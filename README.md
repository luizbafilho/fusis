Fusis Balancer  [![Build Status](https://travis-ci.org/luizbafilho/fusis.svg?branch=master)](https://travis-ci.org/luizbafilho/fusis)
======

Fusis Balancer is a software [Layer 4](https://en.wikipedia.org/wiki/Transport_layer) Load Balancer powered by Linux's [IPVS](http://www.linuxvirtualserver.org/). It is a control plane for IPVS and adds distribution, fault tolerance, self-configuration and a nice JSON API to it.

## Why?
IPVS is hard. Fusis is an abstraction to make it easier to deal with IPVS and make its way to production without problems.

The goal of this project is to provide a friendly way to use IPVS.

## Fault Tolerance
To make sure Fusis does not become a Single Point of Failure in your infrastructure, the Fusis can operate in two modes: `Failover` or `Distributed` modes.

### Failover
In this mode, there is always one single instance balancing the traffic, and `N` others working as secondary instances. Once the Primary is down, a secondary instance becomes the primary and starts balancing the load.

### Distribute
In this mode, all instances balance the traffic. To distribute the traffic to every instance, we need to make use of `ECMP`, so, the router can distribute the traffic equally. Fusis integrates out of the box with BGP without any external dependencies. With a basic configuration, you can peer with your BGP infrastructure and have a distributed load balancer.

```TOML
[bgp]
as = 100
router-id = "192.168.151.182"

  [[bgp.neighbors]]
  address = "192.168.151.178"
  peer-as = 100
```

## State
This project it is under heavy development, it is not usable yet, but you can **Star** :star: the project and follow the updates.

## Dependencies
* Linux kernel >= 2.6.10 or with IPVS module installed
* [libnl 3.X](https://www.infradead.org/~tgr/libnl/)

## Quick Start
WIP

## Documentation

[View documentation →](http://luizbafilho.github.io/fusis/)

## Developing

### VM setup with Vagrant
1. Install [Vagrant](https://www.vagrantup.com)

2. Build the VM
```bash
vagrant up
```
Watch the message at the end of vagrant provision process.
It will provide you with the user, password and where the project code is.

3. Login
```bash
vagrant ssh
```

### Linux setup
1. Install **Go 1.6** or later

2. Install libnl-3 (Debian based: `apt-get install libnl-3-dev libnl-genl-3-dev`)

3. Get this project into $GOPATH:
  ``` bash
  go get -v github.com/luizbafilho/fusis
  ```

### Running the project

Now that you have IPVS and fusis installed, run the project:

``` bash
# Remember, fusis binary is at $GOPATH/bin/fusis, add it to your $PATH
sudo fusis balancer --bootstrap
```
You should see something like:
`[GIN-debug] Listening and serving HTTP on :8000`

From another host, send a HTTP request to the API querying for available services available:
``` bash
curl -i {IP OF FUSIS HOST}:8000/services
```
So you should get:
```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Thu, 07 Apr 2016 21:23:18 GMT
Content-Length: 3

[]
```

Just for testing purposes, lets add a route to a fake IPv4 by running this on the fusis host:

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
