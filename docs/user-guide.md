# 用户使用指南
这篇指南介绍已有系统如何对接网关，以及如何使用网关的一些高级特性

## 已有业务系统
假设有2个业务系统，A和B。A和B两个系统对外提供HTTP服务

|业务系统|地址|
|--|--|
|A-1|192.168.0.101:8080|
|A-2|192.168.0.102:8080|
|B-1|192.168.0.103:8080|
|B-2|192.168.0.104:8080|

### 业务A提供接口

|接口|URL|Method|
|--|--|--|
|查询用户信息|/users/{id}|GET|

返回数据:
```
{
    "code": 0,
    "data": {
        "id": 100,
        "name": "zhangsan",
        "age": 20
    }
}
```

### 业务B提供接口

|接口|URL|Method|
|--|--|--|
|查询用户账户信息|/account/{id}|GET|

返回数据:
```
{
    "code": 0,
    "data": {
        "id": 100,
        "cardNo": "88888888",
        "balance": 100
    }
}
```

## 整合业务系统到网关
网关搭建参见[搭建Gateway环境](./build.md)

### 网关环境信息
|组件|地址|开放端口|
|--|--|--|
|Etcd|190.168.0.10|2379|
|ApiServer|192.168.0.11|9092(grpc),9093(http)|
|Proxy|192.168.0.12|80(http)|

### 网关元数据管理
* [Restful的管理接口](./restful.md)
* [GRPC的管理接口](../examples)

### 网关核心概念
以下是网关的核心概念，请务必理解了再继续阅读
* [Cluster](./cluster.md)
* [Server](./server.md)
* [api](./api.md)

### 创建业务A和B对应的Cluster
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"cluster-A","loadBalance":0}' http://192.168.0.11:9093/v1/clusters

curl -X PUT -H "Content-Type: application/json" -d '{"name":"cluster-B","loadBalance":0}' http://192.168.0.11:9093/v1/clusters
```
记录下对应的返回结果中的ID的字段，用于后续绑定

### 创建业务A和B的真实服务器对应的Server
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.101:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/servers

curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.102:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/servers

curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.103:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/servers

curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.104:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/servers
```
记录下对应的返回结果中的ID的字段，用于后续绑定

### 绑定Server到对应的Cluster
拿到上面步骤创建Cluster和Server元数据的ID的值，绑定Server到对应的Cluster

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":业务A对应的ID,"serverID":192.168.0.101对应的ID}' http://192.168.0.11:9093/v1/binds

curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":业务A对应的ID,"serverID":192.168.0.102对应的ID}' http://192.168.0.11:9093/v1/binds

curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":业务B对应的ID,"serverID":192.168.0.103对应的ID}' http://192.168.0.11:9093/v1/binds

curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":业务B对应的ID,"serverID":192.168.0.104对应的ID}' http://192.168.0.11:9093/v1/binds
```

### 创建业务A的查询用户信息接口到Gateway
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"查询用户信息接口","urlPattern":"^/users/.*$","method":"GET","status":1,"nodes":[{"clusterID":业务A对应的ID}]}' http://192.168.0.11:9093/v1/apis
```

### 创建业务B的查询账户信息接口到Gateway
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"查询账户信息接口","urlPattern":"^/accounts/.*$","method":"GET","status":1,"nodes":[{"clusterID":业务B对应的ID}]}' http://192.168.0.11:9093/v1/apis
```

### 通过访问Gateway访问后端的接口
```bash
curl http://192.168.0.12/users/100

curl http://192.168.0.12/accounts/100
```

## 网关高级特性
### URL重写
比如，现在需要给A业务和B业务的接口添加版本的前缀，在不修改后端接口的情况下，可以利用URL重写来实现：
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"查询用户信息接口2","urlPattern":"^/v1/users/(.*)$","method":"GET","status":1,"nodes":[{"clusterID":业务A对应的ID,"urlRewrite":"/users/$1"}]}' http://192.168.0.11:9093/v1/apis
```

客户端访问接口
```bash
curl http://192.168.0.12/v1/users/100
```

### API聚合
如果一个业务场景同时需要A和B业务的返回数据，并且形成一个新的接口同时返回这些数据，利用Gateway的聚合功能实现：
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"聚合查询接口","urlPattern":"^/aggregation/(.*)$","method":"GET","status":1,"nodes":[{"clusterID":业务A对应的ID,"urlRewrite":"/users/$1","attrName":"user"}, {"clusterID":业务B对应的ID,"urlRewrite":"/accounts/$1","attrName":"account"}]}' http://192.168.0.11:9093/v1/apis
```

客户端访问接口
```bash
curl http://192.168.0.12/v1/aggregation/100
```

返回聚合数据
```json
{
    "user": {
        "code": 0,
        "data": {
            "id": 100,
            "name": "zhangsan",
            "age": 20
        }
    },
    "account": {
        "code": 0,
        "data": {
            "id": 100,
            "cardNo": "88888888",
            "balance": 100
        }
    }
}
```

### 使用路由
加入一个接口使用新的技术重新实现了，上线后，需要分10%的流量到新的实现，可以这样做

创建新版本的Cluster
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"cluster-new-A","loadBalance":0}' http://192.168.0.11:9093/v1/clusters
```

创建新版本对应的Server
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.105:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/server
```

绑定新版本的Server到新的Cluster上
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":业务A新版本的Cluster对应的ID,"serverID":192.168.0.105对应的ID}' http://192.168.0.11:9093/v1/binds
```

创建路由
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":业务A新版本的Cluster对应的ID,"strategy":1,"trafficRate":10,"status":1,"api":原先老接口定义的API的ID,"name":"10%流量导入测试路由"}' http://192.168.0.11:9093/v1/routings
```

### 结果渲染模板
在上面`API聚合`的例子中，返回的结果的格式不符合规范。使用渲染模板调整返回结果，使得返回结果符合规范，修改之前`API聚合创建的API`
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"id":创建返回的ID,"name":"聚合查询接口","urlPattern":"^/aggregation/(.*)$","method":"GET","status":1,"nodes":[{"clusterID":业务A对应的ID,"urlRewrite":"/users/$1","attrName":"user"}, {"clusterID":业务B对应的ID,"urlRewrite":"/accounts/$1","attrName":"account"}],"renderTemplate":{"objects":[{"name":"","attrs":[{"name":"code","extractExp":"user.code"}],"flatAttrs":true},{"name":"data","attrs":[{"name":"user","extractExp":"user.data"},{"name":"account","extractExp":"account.data"}],"flatAttrs":false}]}}' http://192.168.0.11:9093/v1/apis
```

返回数据
```json
{
    "code": 0,
    "data": {
        "user": {
            "id": 100,
            "name": "zhangsan",
            "age": 20
        },
        "account": {
            "id": 100,
            "cardNo": "88888888",
            "balance": 100
        }
    }
}
```

### Mock
利用Gateway创建一个Mock接口。
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"Mock接口","urlPattern":"^/api/mock$","method":"GET","status":1,"defaultValue":{"code":200, "body":"aGVsbG8gd29ybGQ=","headers":["name":"x-mock-header","mock-header-value"],"cookies":["name":"x-mock-cookie","mock-cookie-value"]}}' http://192.168.0.11:9093/v1/apis
```

特别注意，body需要转换为base64编码. `aGVsbG8gd29ybGQ=` 对应 `hello world`

返回Body数据：`hello world`，HTTP Header中包含API中设置的header和cookie设置的值

### Cache结果
对于不经常变化的查询结果，可以在网关缓存，缓解后端压力。

```json
curl -X PUT -H "Content-Type: application/json" -d '{"name":"需要缓存的接口","urlPattern":"^/api/cache$","method":"GET","status":1,"nodes":[{"clusterID":"业务对应的ID","cache":{"deadline":100}}]}' http://192.168.0.11:9093/v1/apis
```

deadline单位是秒，即100s后，cache的值自动清除