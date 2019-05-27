# User Instruction
This instruction aims to explain how existing systems are exposed to the gateway and how to use advanced features of 
the gateway.

## Existing Business Systems
Suppose there are two business systems, A and B, both of which provide external HTTP service.

|Business System|Address|
|--|--|
|A-1|192.168.0.101:8080|
|A-2|192.168.0.102:8080|
|B-1|192.168.0.103:8080|
|B-2|192.168.0.104:8080|

### API of Business A

|API|URL|Method|
|--|--|--|
|query user information|/users/{id}|GET|

Response:
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

### API of Business B

|API|URL|Method|
|--|--|--|
|query user account information|/account/{id}|GET|

Response:
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

## Aggregate Business Systems into The Gateway
Gateway setup reference[Set up gateway environment](./build.md)

### Gateway Environment Information
|Component|IP|Port|
|--|--|--|
|Etcd|190.168.0.10|2379|
|ApiServer|192.168.0.11|9092(grpc),9093(http)|
|Proxy|192.168.0.12|80(http)|

### Gateway Metadata Management
* [Restful Management API](./restful.md)
* [GRPC Management API](../examples)

### Gateway Key Concepts
The following are the key concepts of the gateway. Please make sure you understand them before continuing.
* [Cluster](./cluster.md)
* [Server](./server.md)
* [API](./api.md)

### Create Clusters corresponding to Business System A and B
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"cluster-A","loadBalance":0}' http://192.168.0.11:9093/v1/clusters

curl -X PUT -H "Content-Type: application/json" -d '{"name":"cluster-B","loadBalance":0}' http://192.168.0.11:9093/v1/clusters
```
Record the id field of the responses for later binding.

### Create Servers of Business System A and B
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.101:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/servers

curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.102:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/servers

curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.103:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/servers

curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.104:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/servers
```
Record the id field of the responses for later binding.

### Bind Servers to Corresponding Clusters
Bind servers to corresponding clusters based on the ids from the creation metadata above

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":Business System A-1 ID,"serverID":id of 192.168.0.101}' http://192.168.0.11:9093/v1/binds

curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":Business System A-1 ID,"serverID":id of 192.168.0.102}' http://192.168.0.11:9093/v1/binds

curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":Business System B-1 ID,"serverID":id of 192.168.0.103}' http://192.168.0.11:9093/v1/binds

curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":Business System B-2 ID,"serverID":id of 192.168.0.104}' http://192.168.0.11:9093/v1/binds
```

### Register The User Info Query API of Business System A to The Gateway
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"User Info Query API","urlPattern":"^/users/.*$","method":"GET","status":1,"nodes":[{"clusterID":id of business system A}]}' http://192.168.0.11:9093/v1/apis
```

### Register The User Info Query API of Business System B to The Gateway
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"User Info Query API","urlPattern":"^/accounts/.*$","method":"GET","status":1,"nodes":[{"clusterID":id of business system B}]}' http://192.168.0.11:9093/v1/apis
```

### Access Backend APIs by Sending HTTP requests to The Gateway
```bash
curl http://192.168.0.12/users/100

curl http://192.168.0.12/accounts/100
```

## Advanced Features of The Gateway
### URL (For Client) Rewrite
For instance, suppose there is need to add version prefix to business system A and B, we can do this by rewriting urls without modifying backend APIs:
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"User Info Query API 2","urlPattern":"^/v1/users/(.*)$","method":"GET","status":1,"nodes":[{"clusterID":business system A id,"urlRewrite":"/users/$1"}]}' http://192.168.0.11:9093/v1/apis
```

API Exposed to Client
```bash
curl http://192.168.0.12/v1/users/100
```

### API Aggregation
If a business scenario requires responses from both business system A and B and a new API to handle the aggregated response, this can be achieved by using the gateway's aggregation feature.
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"Query Aggregation API","urlPattern":"^/aggregation/(.*)$","method":"GET","status":1,"nodes":[{"clusterID":id of business A,"urlRewrite":"/users/$1","attrName":"user"}, {"clusterID":id of business B,"urlRewrite":"/accounts/$1","attrName":"account"}]}' http://192.168.0.11:9093/v1/apis
```

API Exposed to Client
```bash
curl http://192.168.0.12/v1/aggregation/100
```

Aggregated Response
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

### Router Usage
Suppose an API is reimplemented using a new tech and 10% of the traffic is routed to it when online, we can do the following.

Create A New Version Cluster
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"cluster-new-A","loadBalance":0}' http://192.168.0.11:9093/v1/clusters
```

Create The Corresponding Server
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"addr":"192.168.0.105:8080","protocol":0,"maxQPS":100}' http://192.168.0.11:9093/v1/server
```

Bind The New Version Server to The New Version Cluster
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":id of the new version business A cluster,"serverID":id of 192.168.0.105}' http://192.168.0.11:9093/v1/binds
```

Create Router
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"clusterID":id of the new version business A cluster,"strategy":1,"trafficRate":10,"status":1,"api":the id of the previous API,"name":"10% of the traffic goes to the new cluster"}' http://192.168.0.11:9093/v1/routings
```

### Result Rendering Template
In the above API Aggregation example, the response is not in compliance with the standard. We can use the rendering template to adjust the response to make it be in compliance.
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"id":id of the api (found in response from creation),"name":"Aggregated Query API","urlPattern":"^/aggregation/(.*)$","method":"GET","status":1,"nodes":[{"clusterID":id of business system A,"urlRewrite":"/users/$1","attrName":"user"}, {"clusterID":id of business system B,"urlRewrite":"/accounts/$1","attrName":"account"}],"renderTemplate":{"objects":[{"name":"","attrs":[{"name":"code","extractExp":"user.code"}],"flatAttrs":true},{"name":"data","attrs":[{"name":"user","extractExp":"user.data"},{"name":"account","extractExp":"account.data"}],"flatAttrs":false}]}}' http://192.168.0.11:9093/v1/apis
```

Response
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
Use Gateway to Create A Mock API
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"Mock API","urlPattern":"^/api/mock$","method":"GET","status":1,"defaultValue":{"code":200, "body":"aGVsbG8gd29ybGQ=","headers":["name":"x-mock-header","mock-header-value"],"cookies":["name":"x-mock-cookie","mock-cookie-value"]}}' http://192.168.0.11:9093/v1/apis
```

Attention, body field requires base64 encoding. `aGVsbG8gd29ybGQ=` maps to `hello world`

Body response：`hello world`. HTTP Header includes API's already-set header and cookie.

### Cache Result
For those pretty much consistent query result, cache in the gateway is allowed to alleviate backend pressure.

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"API in need of cache","urlPattern":"^/api/cache$","method":"GET","status":1,"nodes":[{"clusterID":"id of business system","cache":{"deadline":100}}]}' http://192.168.0.11:9093/v1/apis
```

The deadline unit of time is in seconds. In the above example, cache is purged automatically after 100 seconds.