Filter javascript plugin
--------------
Gateway提供以`Javascript`编写插件的能力，用以动态的扩展功能，甚至可以提供`Serverless`的能力扩展功能。

## 定义
一个js的插件定义如下：
```javascript
function NewPlugin(cfg) {
    // import builtin modules
    var JSON = require("json")
    var HTTP = require("http")
    var REDIS = require("redis")
    var LOG = require("log")

    // init plugin
    return {
        // filter pre method
        "pre": function(ctx) {
            // biz code
            // ...
            // biz code
            return {
                "code": 200,
                "error": "error",
            }
        },
        // filter post method
        "post": function(ctx) {
            // biz code
            // ...
            // biz code
            return {
                "code": 200,
                "error": "error",
            }
        },
        "postErr": function(ctx) {
            // biz code
            // ...
            // biz code
        }
    }
}
```

## ctx 插件上下文对象
|method|args|return|remark|
| - | - | - | - |
|OriginRequest||HTTPRequest|接收到的原始请求|
|ForwardRequest||HTTPRequest|转发到后端的请求|
|Response||HTTPResponse|后端响应|
|SetAttr|key String, value Any Type||在上下文中存储属性，用来在插件之间传递数据|
|HasAttr|key String|Boolean|检查一个属性是否在上下文中存在
|GetAttr|key String|Any Type|获取上下文中的属性|

### HTTPRequest对象
|method|args|return|remark|
| - | - | - | - |
|Header|name String|String|获取请求header|
|RemoveHeader|name String||移除请求Header|
|SetHeader|name String, value String||设置请求Header|
|Cookie|name String|String|获取请求Cookie|
|RemoveCookie|name String||移除请求Cookie|
|SetCookie|name String, value String||设置请求Cookie|
|Query|name String|String|获取请求URL参数|
|Body||String|获取请求Body|
|SetBody|String||设置请求Body|

### HTTPResponse对象
|method|args|return|remark|
| - | - | - | - |
|Header|name String|String|获取响应header|
|RemoveHeader|name String||移除响应Header|
|SetHeader|name String, value String||设置响应Header|
|Cookie|name String|String|获取响应Cookie|
|RemoveCookie|name String||移除响应Cookie|
|SetCookie|domain String, path String, name String, value String, expire int64, httpOnly Boolean, secure Boolean||设置响应Cookie|
|Query|name String|String|获取响应URL参数|
|Body||String|获取响应Body|
|SetBody|String||设置响应Body|
|SetStatusCode|Integer||设置响应状态码

## 内建模块
由于js执行引擎并不兼容nodejs module，所以不能使用nodes module，所以gateway提供了一些内建的模块帮助编写插件。`由于这些内建库都是GO实现的，所以所有的方法名称首字母都是大写`，这里的习惯和js有点违背，但无关大雅

### JSON
JSON内建插件JSON的编解码

|method|args|return|remark|
| - | - | - | - |
|Stringify|JSON|String||
|Parse|String|JSON||

### LOG
提供插件日志打印到gateway的日志文件中

|method|args|return|remark|
| - | - | - | - |
|Info|String|||
|Infof|formart String, args ...AnyType|||
|Debug|String|||
|Debugf|formart String , args ...AnyType|||
|Warn|String|||
|Warnf|formart String, args ...AnyType |||
|Error|String|||
|Errorf|formart String, args ...AnyType |||
|Fatal|String||proxy进程会退出|
|Fatalf|formart String, args ...AnyType ||proxy进程会退出|

### HTTP
提供插件HTTP的能力，并提供同步的编程模式，返回结果就是HTTP的响应

|method|args|return|remark|
| - | - | - | - |
|Get|String|HTTPResult||
|Post|url String, body string, header JSON|HTTPResult|需要自己在header中设置Content-Type|
|PostJSON|url String, body string, header JSON|HTTPResult|会自动在header中设置Content-Type的类型为application/json|
|Put|url String, body string, header JSON|HTTPResult|需要自己在header中设置Content-Type|
|PutJSON|url String, body string, header JSON|HTTPResult|会自动在header中设置Content-Type的类型为application/json|
|Delete|url String, body string, header JSON|HTTPResult|需要自己在header中设置Content-Type|
|DeleteJSON|url String, body string, header JSON|HTTPResult|会自动在header中设置Content-Type的类型为application/json|

#### HTTPResult
|method|args|return|remark|
| - | - | - | - |
|HasError||Boolean|判断本次HTTP请求有没有错误|
|Error||String|返回本次HTTP请求错误|
|StatusCode||Number|返回本次请求的响应状态码|
|Header||Object|返回响应的Header|
|Cookie||Object Array|返回响应的Cookie|
|Body||String|响应Body体|

### Redis
提供js访问Redis的能力

|method|args|return|remark|
| - | - | - | - |
|RedisModule|Object|RedisOp|创建Redis操作对象，`{"maxActive":"最大链接数，int", "maxIdle":"链接最大idle链接数，int", "idleTimeout":"idle超时时间，单位秒,int", "addr":"redis地址"}`|

### RedisOp

|method|args|return|remark|
| - | - | - | - |
|Do|cmd String, args ...Any Type|CmdResp|执行redis命令|

### CmdResp

|method|args|return|remark|
| - | - | - | - |
|HasError||Boolean|检查本次操作是否有错误|
|Error||String|返回本次操作的错误|
|StringValue||String|把本次操作的结果转换为String|
|StringsValue||String Array|把本次操作的结果转换为String Array|
|StringMapValue||Object|把本次操作的结果转换为JSON，一般用于hash结构|
|IntValue||Integer|把本次操作的结果转换为Integer|
|IntsValue||Integer Array|把本次操作的结果转换为Integer Array|
|IntMapValue||Object|把本次操作的结果转换为JSON ，一般用于hash结构|
|Int64Value||Long|把本次操作的结果转换为Long|
|Int64sValue||Long Array|把本次操作的结果转换为Long Array|
|Int64MapValue||Object|把本次操作的结果转换为JSON ，一般用于hash结构|