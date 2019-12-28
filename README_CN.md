<img src="./images/logo.png" height=80></img>

[![Gitter](https://badges.gitter.im/fagongzi/gateway.svg)](https://gitter.im/fagongzi/gateway?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Build Status](https://api.travis-ci.org/fagongzi/gateway.svg)](https://travis-ci.org/fagongzi/gateway)
[![Go Report Card](https://goreportcard.com/badge/github.com/fagongzi/gateway)](https://goreportcard.com/report/github.com/fagongzi/gateway)

Manba/[English](./README.md)
-------
Manba是一个基于HTTP协议的restful的API网关。可以作为统一的API接入层。

## 教程
如果你是一个初学者，那么这个[详细的教程](./docs/tutorial.md)非常适合你。现在只有英文版本。

## 注意
请确保你的Go版本是在1.10或者之上。用1.10之前版本的Go编译时会出现**undefined "math/rand".Shuffle**错误。[StackOverFlow链接](https://stackoverflow.com/questions/52172794/getting-undefined-rand-shuffle-in-golang)

## Features
* 流量控制(Server或API级别)
* 熔断(Server或API级别)
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
* JWT Authorization
* API Metric导入Prometheus
* API 失败重试
* 后端server的健康检查
* 开放管理API(GRPC、Restful)
* 支持websocket
* 支持在线迁移数据

## Docker

以下内容要求对docker基本操作有一定了解，可以看[这本书][2]，或者直接看[官方文档][1]。

### 快速开始
使用 `docker pull fagongzi/manba` 命令下载Docker镜像, 使用 `docker run -d -p 9093:9093 -p 80:80 -p 9092:9092 fagongzi/manba` 运行镜像. 镜像启动后export 3个端口:

* 80

  Proxy的http端口，这个端口就是直接为终端用户服务的

* 9092

  APIServer的对外GRPC的端口

* 9093

  APIServer的对外HTTP Restful的端口，访问 `http://127.0.0.1:9093/ui/index.html`访问WEBUI

通过设置以下环境变量可以改变运行参数，参数相同时配置参数将会覆盖默认参数

- GW_PROXY_OPTS

   支持`proxy --help`中的所有参数；

- API_SERVER_OPTS

   支持`apiserver --help`中的所有参数；

- ETCD_OPTS

   支持`etcd --help`中的所有参数；

### 可用的docker镜像

* `fagongzi/proxy`

   proxy组件，`生产可用`

* `fagongzi/apiserver`

   apiserver组件，`生产可用`

### Quick start with docker-compose
```bash
docker-compose up -d
```

使用 `http://127.0.0.1:9093/ui/index.html` 访问 `apiserver`

使用 `http://127.0.0.1` 访问你的API

## 架构
![](./images/arch.png)

## WebUI
可用的Manba的WebUI的项目：
* [官方](https://github.com/fagongzi/gateway-ui-vue)
* [gateway_ui（仅适配2.x）](https://github.com/archfish/gateway_ui)
* [gateway_admin_ui](https://github.com/wilehos/gateway_admin_ui)

## 组件
Gateway由`proxy`, `apiserver`组成

### Proxy
Proxy是Gateway对终端用户提供服务的组件，Proxy是一个无状态的节点，可以部署多个来支撑更大的流量，[更多](./docs-cn/proxy.md)。

### ApiServer
ApiServer对外提供GRPC和Restful来管理元信息，ApiServer同时集成了官方的WebUI，[更多](./docs-cn/apiserver.md)。

## Manba中的概念
### Server
Server是一个真实的后端服务，[更多](./docs-cn/server.md)。

### Cluster
Cluster是一个逻辑概念，它由一组提供相同服务的Server组成。会依据负载均衡策略选择一个可用的Server，[更多](./docs-cn/cluster.md)。

### API
API是Manba的核心概念，我们可以在Manba的中维护对外的API，以及API的分发规则，聚合规则以及URL匹配规则，[更多](./docs-cn/api.md)。

### Routing
Routing是一个路由策略，根据HTTP Request中的Cookie，Querystring、Header、Path中的一些信息把流量分发到或者复制到指定的Cluster，通过这个功能，我们可以实现AB Test和线上引流，[更多](./docs-cn/routing.md)。

## 参与开发
[更多](./docs-cn/build.md)

## 交流方式-微信
![](./images/qr.jpg)

[1]: https://docs.docker.com/ "Docker Documentation"
[2]: https://github.com/yeasy/docker_practice "docker_practice"
