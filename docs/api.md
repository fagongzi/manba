API
-----------
API是Gateway的核心概念。可以通过Gateway的API-Server管理API。

# API属性
## ID
API的ID，唯一标识一个API。

## Name
API的名称。

## URLPattern
URL匹配模式，使用正则表达式表示。Gateway使用该字段来匹配原始请求的URL。该字段必须和`Method`配合使用，同时满足才算这个请求匹配了这个API。

## Method
HTTP Method， `*` 匹配所有的HTTP Method（GET,PUT,POST,DELETE）。该字段必须和`URLPattern`配合使用，同时满足才算这个请求匹配了这个API。

## Domain（可选）
host，当原始请求的host等于该值，则认为匹配了当前的API，同时忽略`URLPattern`和`Method`。

## Status
API 状态枚举, 有2个值组成： `UP` 和 `Down`。只有`UP`状态才能生效。

## IPAccessControl（可选）
IP的访问控制，有黑白名单2个部门组成。

## DefaultValue（可选）
API的默认返回值，当后端Cluster无可用Server的时候，Gateway将返回这个默认值，默认值由HTTP Body、Header、Cookie三部分组成。改值可以用来做Mock。
  
## Nodes
请求被转发到的后端Cluster。至少设置一个转发Cluster，一个请求可以被同时转发到多个后端Cluster（目前仅支持GET请求设置多个转发）。在转发的时候，针对每一个转发支持以下特性：

* 支持URL重写

  例如，API对外提供的URL是`/api/users/1`，后端真实server提供的URL是`/users?id=1`，类似这种情况需要对原始URL进行重写。
  对于这个重写，我们需要配置API的`URLPattern`属性为`/api/users/(\d+)`，并且配置转发的URL重写规则为：`users?id=$1`
* 支持对原始请求的参数校验
  
  支持针对`querystring`、`json body`、`cookie`、`header`中的任意属性配置正则表达式的校验规则
* 聚合多个后端Cluster的响应，统一返回

  支持一个请求被同时分发到多个后端Cluster，并且为每一个后端Cluster返回的数据设置一个属性名，并且聚合所有的返回值作为一个JSON统一返回。例如：一个前端APP的页面需要显示用户账户信息以及用户的基本信息，可以使用这个特性，定制一个API`/api/users/(\d+)`，同时配置分发到2个后端Cluster，并且配置URL的重写规则为`/users/base/$1`和`/users/account/$1`，这样聚合2个信息返回。

## Perms
设置访问这个API需要的权限，需要用户自己开发权限检查插件。

## AuthFilter
指定该API所使用的Auth插件名称，Auth插件的实现可以借鉴[JWT插件](https://github.com/fagongzi/jwt-plugin)