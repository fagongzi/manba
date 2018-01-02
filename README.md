<img src="./images/logo.png" height=80></img>

[![Gitter](https://badges.gitter.im/fagongzi/gateway.svg)](https://gitter.im/fagongzi/gateway?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Build Status](https://api.travis-ci.org/fagongzi/gateway.svg)](https://travis-ci.org/fagongzi/gateway)
[![Go Report Card](https://goreportcard.com/badge/github.com/fagongzi/gateway)](https://goreportcard.com/report/github.com/fagongzi/gateway)

Gateway
-------
Gateway 是一个基于HTTP协议的restful的API网关。可以作为统一的API接入层。

# Note
原先老版本的Gateway不在维护，这个master是基于2.0.0版本，从这个版本开始，不再支持Admin管理UI，不再支持Consul作为元数据存储。元数据的管理提供客户端来管理。可以使用[元数据迁移工具](https://github.com/fagongzi/migrater) 迁移元数据(Cluster、Server、Bind、API)到新版本的gateway上。

# Features
* 流量控制
* 熔断
* 负载均衡
* 服务发现
* 插件机制
* 路由
* API 聚合
* API 参数校验
* API 访问控制（黑白名单）
* API 默认返回值
* 后端server的健康检查
* 使用 [fasthttp](https://github.com/valyala/fasthttp)
* 开放管理API

# Install
[更多](./docs/build.md)

# Docker
使用 `docker pull fagongzi/gateway` 命令下载Docker镜像, 使用 `docker run -d fagongzi/gateway` 运行镜像. 镜像启动后export 3个端口:

* 80

  Proxy的http端口，这个端口就是直接为终端用户服务的

* 9092

  APIServer的对外GRPC的端口

# 架构
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
Routing是一个路由策略，根据HTTP Request中的cookie，query string、header中的一些信息把流量分发到指定的Cluster，通过这个功能，我们可以实现AB Test，[更多](./docs/routing.md)。
  
# Gateway能帮助你做什么
## API 定义
你可以在Gateway中动态的定义API，这些API可以对应后端服务的一个真实的API或者一组真实的API。并且可以随时上线和下线这些API。

## Validation
你可以为每一个API设置一组校验规则，这些校验规则可以动态的修改，并且实时生效。

## API Mock
你可以完全Mock一个后端不存在的API或者为一个API作为发生错误的返回默认值，定义返回的JSON数据以及Header、Cookie等信息。使用这个功能可以用来作为后端服务错误（被流控，被降级等）的时候的默认返回值；也可以帮助前端开发人员在开发阶段独立完成功能。

## 保护后端服务
Gateway 通过 **流控** and **熔断** 的功能保护后端服务。

## AB Test
Gateway 通过 **Routing** 功能可以做AB测试。

## 线上引流
Gateway 通过 **Routing** 功能可以线上引流。

# 插件机制
Gateway以go1.8的plugin机制提供如下的扩展点

* filter
  使用go1.8的plugin的机制，编写自定义插件，扩展gateway功能。[如何编写自定义filter](./docs/plugin.md)

# 交流方式
微信: 13675153174