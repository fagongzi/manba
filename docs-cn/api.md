API
-----------
API是Gateway的核心概念. 你可以通过Admin组件管理你的API。

# API属性
## Name
API的名称，唯一标示一个API。

## URL
API的URL，URL定义是一个正则表达式，当一个请求的URL匹配时，Proxy会把这个请求转发到这个API关联的分发节点上。

## Method
API HTTP Method， 必须同时匹配URL和Method才算是匹配了这个API. `*` 匹配所有的HTTP Method（GET,PUT,POST,DELETE）

## Status
API status, 有2个值组成： `UP` 和 `Down`。一个API在UP状态才能被使用。

## 访问控制
访问控制是一组白名单或者黑名单的规则，通过客户端的IP来匹配。这个配置是一个JSON格式的：
  
```json
{
    "blacklist": [
        "127.*",
        "127.0.0.*",
        "127.0.*",
        "127.0.0.1"
    ],
    "whitelist": [
        "127.*",
        "127.0.0.*",
        "127.0.*",
        "127.0.0.1"
    ]
}
```

## Mock
  Mock是当API对应的真实后端服务或者后端服务出错的情况下的Mock返回数据，它是一个JSON配置：

```json
{
"value": "{\"abc\":\"hello\"}",
"contentType": "application/json; charset=utf-8",
"headers": [
    {
        "name": "header1",
        "value": "value1"
    }
],
"cookies": [
    "test-c=1",  
    "test-c2=2"  
]
}
```
其中value字段是必选字段, contentType, headers 和 cookies是可选字段。
  
## Nodes
在Gateway中，可以为每个API设置一个或多个分发节点，Proxy会把原始的Request请求同时分发到这些节点去，同时等待这些节点的响应并且合并这个响应结果，然后给客户端响应。每个节点包含4个属性：Cluster、Attrbute name、Rewrite和Validations

### Cluster （必填）
原始请求被分发到哪个Cluster，Cluster接收到请求后，会根据设置的负载均衡算法选择一个Server发送请求。

### Attrbute name (在多于一个节点的时候必填)
当我们设置了多个分发节点，Proxy会吧返回的结果合并成一个JSON，这里的Attrbute name是设置每个分发节点的返回值在这个JSON中的属性名称。例如：设置两个分发节点Node1和Node2，Node1的Attrbute name设置为`base`，Node1的Attrbute name设置为`account`，那么最后合并后的结果为：`{"base": {"name":"user1"}, "account": {"money":1000}}`

### Rewrite (可选项)
URL Rewrite规则用于重写原始URL，这个参数通常和API的`URL`属性配合。一个常见的场景，后端的Server提供的API的URL可能并不是Restful的接口，然而我们期望在定义一个Restful的API，那么就会存在URL的转换需求。举个例子：后端真实Server提供的接口为`/users?userId=1`，我们期望提供`/users/1`的API，我们可以设置API的`URL`属性为`/api/users/(.+)`，Node的`Rewrite`规则为`/users?userId=$1`，这里利用了正则表达式**捕获型括号**的机制来提供URL的重写。

### Validations (可选项)
Validations 规则用来校验请求参数是否符合预期。这是一个JSON配置：

```json
[
    {
        "attr": "abc",  
        "getFrom": 0,   // enum value, 0: query string. 1: form data 
        "required": true, 
        "rules": [
            {
                "type": 0, 
                "expression": "\\d+" 
            },
            {
                "type": 0,
                "expression": "\\d+"
            }
        ]
    }
]
``` 

这是一个JSON数组，每一个JSON元素是一个参数的验证规则，包含4个属性：
* attr 
  属性名称
* getFrom 
    对应`attr`参数，枚举值：0：query string参数；1：form data 参数
* required
    是否是必选参数
* rules
    校验规则集合，目前只支持正则校验规则，`type`字段必须设置为0。
