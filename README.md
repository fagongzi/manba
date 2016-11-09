<img src="images/logo.png" height=80></img>

[![Gitter](https://badges.gitter.im/fagongzi/gateway.svg)](https://gitter.im/fagongzi/gateway?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Build Status](https://api.travis-ci.org/fagongzi/gateway.svg)](https://travis-ci.org/fagongzi/gateway)
[![Go Report Card](https://goreportcard.com/badge/github.com/fagongzi/gateway)](https://goreportcard.com/report/github.com/fagongzi/gateway)

Gateway
-------
Gateway is a API gateway based on http. It works at 7 layer.

# Features
* Traffic Control
* Circuit Breaker
* Loadbalance
* Routing based on URL
* API aggregation(support url rewrite)
* API Validation
* API Access Control(blacklist and whitelist)
* API Mock
* Backend Server heath check
* Use [fasthttp](https://github.com/valyala/fasthttp)
* Admin WEBUI

# Install
Gateway dependency [etcd](https://github.com/coreos/etcd)

## Compile from source
```
git clone https://github.com/fagongzi.git
cd $GOPATH/src/github.com/fagongzi/gateway
go build cmd/proxy/proxy.go
go build cmd/admin/admin.go
```

## Download binary file
[linux-64bit](http://7xtbpp.com1.z0.glb.clouddn.com/gateway-linux64.tar.gz)

You can read [this](./docs/build.md) for more infomation about build and run gateway.

## Docker
You can run `docker pull fagongzi/gateway` to get docker images, then use `docker run -d fagongzi/gateway` to run. It export 3 ports:

* 80

  proxy http serve port

* 8081

  proxy manager port

* 8080
  
  admin http port

# Online Demo

* admin

  http://demo-admin.fagongzi.win admin/admin

* proxy
  
  http://demo-proxy.fagongzi.win 

# Architecture
![](./images/arch.png)

## Components
Gateway has three component: proxy, admin, etcd.

### Proxy
The proxy provide http server. Proxy is stateless, you can scale proxy node to deal with large traffic.

### Admin 
The admin is a backend manager system. Admin also is a stateless node, you can use a Nginx node for HA. One Admin node can manager a set of proxy which has a same etcd prefix configuration.

### Etcd
The Etcd store gateway's mete data.

## Concept of gateway

* Server

  Server is a backend server which provide restfule json service.The server is the basic unit at gateway. [more](./docs/server.md).

* Cluster

  Cluster is a set of servers which provide the same service. The Loadbalancer select a usable server to use. [more](./docs/cluster.md).

* API

  API is the core concept in gateway.  You can define a API with a URL pattern, http method, and at least one dispatch node. [more](./docs/api.md).

* Routing

  Routing is a approach to control http traffic to clusters. You can use cookie, query string, request header infomation in a expression for control.

# What gateway can help you
## Redefine your API URL
Your backend server provide some restful API, You can redefine the API URL that provide to API caller.Use this funcation you can provide beautiful APIs.  

## Dynamic URL & Aggregation
You can define a URL and configuration a URL set which you want to aggregation.

## Protect backend server
Gateway can use **Traffic Control** and **Circuit Breaker** functions to avoid backend crash by hight triffic.

## AB Test
Gateway's **Routing** fucntion can help your AB Test.
