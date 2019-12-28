API
-----------
API是Manba的核心概念。可以通过Manba的API-Server管理API。

# API属性
## ID
API的ID，唯一标识一个API。

## Name
API的名称。

## URLPattern
URL匹配表达式，Manba使用该字段来匹配原始请求的URL。该字段必须和`Method`配合使用，同时满足才算这个请求匹配了这个API。

### URLPattern表达式
定义API的`URL`，使用`/`来分割URL Path的每个部分，每个部分可以这些类型：

* 常量字符串 任意URL合法的字符串，可以使用`*`匹配任何字符串
* (number):argeName 指定这个部分是一个数字变量
* (string) 指定这个部分是一个字符串变量
* (enum:enum1|enum1|enum1) 指定这部分是一个枚举变量，可选的枚举值使用|分割
* 如果需要使用这些变量（例如在URL Rewrite的时候）可以使用:
    * (number):id    变量名称为id
    * (string):name  变量名称为name
    * (enum:on|off):action  变量名称为action

一些例子：
* `/api/v1/product/(number):id`
* `/api/v1/product/(number):id/(enum:online|offline):action`
* `/api/v1/*`

## Method（可选）
匹配请求的`HTTP Method`字段， `*` 匹配所有的HTTP Method（GET,PUT,POST,DELETE）。

## Domain（可选）
匹配请求的HOST字段。

## MatchRule
`URLPattern`,`Method`,`Domain`的匹配规则
* MatchDefault `URLPattern` && (`Method` || `Domain`)
* MatchAll `URLPattern` && (`Method` && `Domain`)
* MatchAny `URLPattern` && (`Method` || `Domain`)

## Status
API 状态枚举, 有2个值组成： `UP` 和 `Down`。只有`UP`状态才能生效。

## IPAccessControl（可选）
IP的访问控制，有黑白名单2个部门组成。

## DefaultValue（可选）
API的默认返回值，当后端Cluster无可用Server的时候，Manba将返回这个默认值，默认值由Code、HTTP Body、Header、Cookie组成。可以用来做Mock或者后端服务故障时候的默认返回。

## 聚合请求
聚合请求是原始请求转发到多个后端Server，并且把多个返回结果合并成一个JSON返回，并且可以指定每个转发请求的结果在最终JSON对象中的属性名称。多个转发请求可以同时发送，也可以按批次发送(后面请求参数依赖前面请求的返回值)。使用`BatchIndex`来设置每个转发请求的顺序。

例子
* 原始请求: `/api/v1/aggregation/1`
* 转发请求: `/api/v1/users/1`，返回：`{"name":"zhangsan"}`，属性名为`user`
* 转发请求: `/api/v1/accounts/1`，返回：`{"type":"test", "accountId":"123"}`，属性名为`account`
* 最终返回结果为：`{"user":{"name":"zhangsan"}, "account":{"type":"test", "accountId":"123"}}`

## Nodes
请求被转发到的后端Cluster。至少设置一个转发Cluster，一个请求可以被同时转发到多个后端Cluster（目前仅支持GET请求设置多个转发）。在转发的时候，针对每一个转发支持以下特性：

### 支持URL重写
使用URL重写表达式来重构转发到后端Server的真实URL

#### URL重写表达式
定义重写的`URL`，支持在表达式中使用变量，并且使用运行期用真实的值替换这些变量，变量使用`$()`包裹，用`.`来表示变量的属性

##### origin变量
origin变量用来提取原始请求的`query string`,`header`,`cookie`,`body`中的数据，对于query中的多个同名参数，不支持独立获取

假设原始请求
* URL: `/api/v1/users?id=1&name=fagongzi&page=1&pageSize=100`,
* Header: `x-test-token: token1, x-test-id: 100`
* Cookie: `x-cookie-token: token2, x-cookie-id: 200`
* Body: `{"type":1, "value":{"id":100, "name":"zhangsan"}}`

变量例子
* $(origin.query) = `?id=1&name=fagongzi&page=1&pageSize=100`
* $(origin.path) = `/api/v1/users`
* $(origin.query.id) = `1`, $(origin.query.name) = `fagongzi`, $(origin.query.page) = `1`, $(origin.query.pageSize) = `100`
* $(origin.header.x-test-token) = `token1`, $(origin.header.x-test-id) = `100`
* $(origin.cookie.x-cookie-token) = `token2`, $(origin.cookie.x-cookie-id) = `200`
* $(origin.body.type) = `1`, $(origin.body.value.id) = `100`, $(origin.body.value.name) = `zhangsan`

#### param变量
param变量用来提取API的`URLPattern`中定义的变量

假设:
* API的`URLPattern`为`/api/v1/users/(number):id/(enum:on|off):action`
* 请求的URL为`/api/v1/users/100/on`

变量例子
* $(param.id) = `100`
* $(param.action) = `on`

#### depend变量
depend变量用来提取依赖聚合请求返回值中的数据

假设聚合请求有2个,分别为user和account
* user: {"name":"zhangsan"}
* account: {"id":"123456"}

变量例子
* $(depend.user.name) = `zhangsan`
* $(depend.account.id) = `123456`

#### 一些URL重写的表达式例子
* `/api/v1/users$(origin.query)`
* `$(origin.path)?name=$(origin.header.x-user-name)&id=$(origin.body.user.id)`
* `/api/v1/users?id=$(param.id)&action=$(param.action)`
* `/api/v1/accounts?id=$(depend.user.accountId)`

### 支持对原始请求的参数校验
  支持针对`querystring`、`json body`、`cookie`、`header`、`path value`中的任意属性配置正则表达式的校验规则

###  支持失败重试
  可以设置`retryStrategy`指定根据http返回码重试请求，可以设置重试最大次数以及重试间隔。

### 支持API级别的超时时间覆盖全局设置
  可以设置`ReadTimeout`和`WriteTimeout`来指定请求的读写超时时间，不设置默认使用全局设置。

## Perms（可选）
设置访问这个API需要的权限，需要用户自己开发权限检查插件。

## AuthFilter（可选）
指定该API所使用的Auth插件名称，Auth插件的实现可以借鉴[JWT插件](https://github.com/fagongzi/jwt-plugin)

## RenderTemplate
使用RenderTemplate可以重新定义返回的数据，包括数据的格式，字段等等。

## UseDefault（可选）

当该值为True且`DefaultValue`存在时，直接使用`DefaultValue`作为返回值。

## MatchRule（可选）

| MatchRule | Logic |
| - | - |
| MatchDefault | `Domain` \|\| (`URLPattern` && `Method`) |
| MatchAll | `Domain` && `URLPattern` && `Method` |
| MatchAny | `Domain` \|\| `URLPattern` \|\| `Method` |

## Position（可选）

API匹配时按该值的升序匹配，即值越小优先级越高。默认值为0。

## Tags（可选）
给API加上Tag标签，便于维护和检索。

## WebSocketOptions（可选）
websocket选项，设置该API为`websocket`，注意：`websocket特性还处于试验阶段，默认关闭，可以使用--websocket启用特性`。网关转发websocket的时候，`Origin`默认使用后端Server的地址，如果需要设置特殊值，可以指定`Origin`参数。

## MaxQPS（可选）
API能够支持的最大QPS，用于流控。Manba采用令牌桶算法，根据QPS限制流量，保护后端API被压垮。API的优先级高于`Server`的配置

## CircuitBreaker（可选）
熔断器，设置后端API的熔断规则，API的优先级高于`Server`的配置。熔断器分为3个状态：

* Open

  Open状态，正常状态，Manba放入全部流量。当Manba发现失败的请求比例达到了设置的规则，熔断器会把状态切换到Close状态

* Half

  Half状态，尝试恢复的状态。在这个状态下，Manba会尝试放入一定比例的流量，然后观察这些流量的请求的情况，如果达到预期就把状态转换为Open状态，如果没有达到预期，恢复成Close状态

* Close

  Close状态，在这个状态下，Manba禁止任何流量进入这个后端Server，在达到指定的阈值时间后，Manba自动尝试切换到Half状态，尝试恢复。
