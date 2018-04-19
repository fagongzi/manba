<img src="./images/logo.png" height=80></img>

[![Gitter](https://badges.gitter.im/fagongzi/gateway.svg)](https://gitter.im/fagongzi/gateway?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Build Status](https://api.travis-ci.org/fagongzi/gateway.svg)](https://travis-ci.org/fagongzi/gateway)
[![Go Report Card](https://goreportcard.com/badge/github.com/fagongzi/gateway)](https://goreportcard.com/report/github.com/fagongzi/gateway)

Gateway
-------
Gateway 是一个基于HTTP协议的restful的API网关。可以作为统一的API接入层。

## Features
* 流量控制
* 熔断
* 负载均衡
* 服务发现
* 插件机制
* 路由(分流，复制流量)
* API 聚合
* API 参数校验
* API 访问控制（黑白名单）
* API 默认返回值
* API 定制返回值
* API 结果Cache
* 后端server的健康检查
* 开放管理API(GRPC、Restful)

## Docker
使用 `docker pull fagongzi/gateway` 命令下载Docker镜像, 使用 `docker run -d fagongzi/gateway` 运行镜像. 镜像启动后export 3个端口:

* 80

  Proxy的http端口，这个端口就是直接为终端用户服务的

* 9092

  APIServer的对外GRPC的端口

## 架构
![](./images/arch.png)

## 组件
Gateway由`proxy`, `apiserver`组成

### Proxy
Proxy是Gateway对终端用户提供服务的组件，Proxy是一个无状态的节点，可以部署多个来支撑更大的流量，[更多](./docs/proxy.md)。

### ApiServer 
ApiServer对外提供GRPC的接口，用来管理元信息，[更多](./docs/apiserver.md)。

## Gateway中的概念
### Server
Server是一个真实的后端服务，[更多](./docs/server.md)。

### Cluster
Cluster是一个逻辑概念，它由一组提供相同服务的Server组成。会依据负载均衡策略选择一个可用的Server，[更多](./docs/cluster.md)。

### API
API是Gateway的核心概念，我们可以在Gateway的中维护对外的API，以及API的分发规则，聚合规则以及URL匹配规则，[更多](./docs/api.md)。

### Routing
Routing是一个路由策略，根据HTTP Request中的Cookie，Querystring、Header、Path中的一些信息把流量分发到或者复制到指定的Cluster，通过这个功能，我们可以实现AB Test和线上引流，[更多](./docs/routing.md)。
  
# 交流方式-微信
![](./images/qr.jpg)
