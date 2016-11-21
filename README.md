Fusis Balancer  [![Build Status](https://travis-ci.org/luizbafilho/fusis.svg?branch=master)](https://travis-ci.org/luizbafilho/fusis)
======

Fusis Balancer is a software [Layer 4](https://en.wikipedia.org/wiki/Transport_layer) Load Balancer powered by Linux's [IPVS](http://www.linuxvirtualserver.org/). It exposes a HTTP API to manage your services dynamically.

A layer 3 balancer take decisions based only on IP address (source and destination), a layer 4 balancer can also see transport information like TCP and UDP ports. Being a software balancer, it's tailored to be easy to deploy and scale.

## Why?
The goal of this project is to provide a friendly way to use IPVS.

It will be responsible for detecting new/failed nodes and add/remove routes to them automatically configuring the network to do so.

## State
This project it's under heavy development, it's not usable yet, but you can **Star** :star: the project and follow the updates.

## Dependencies
* Linux kernel >= 2.6.10 or with IPVS module installed
* [libnl 3.X](https://www.infradead.org/~tgr/libnl/)

## Quick Start
WIP

## Developing

### VM setup with Vagrant
1. Install [Vagrant](https://www.vagrantup.com)

2. Build the VM
```bash
vagrant up
```
Watch the message at the end of vagrant provision process.
It'll provide you with user, password and where the project code is.

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

### Logging
Fusis uses [Logrus](https://github.com/Sirupsen/logrus) as its logging system.
By default, Fusis logs to stdout every minute.
You can change its log collection interval by passing the following command line argument:

```bash
# The argument --log-interval or -i. The value is in seconds
sudo fusis balancer --bootstrap --log-interval 10
 ```
