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

### ctx Plugin Context Object
|method|args|return|remark|
| - | - | - | - |
|OriginRequest||HTTPRequest|Original Request Received|
|ForwardRequest||HTTPRequest|Requests Redirected to Backend|
|Response||HTTPResponse|Backend Response|
|SetAttr|key String, value Any Type||Used to store attributes and transmitt data between plugins|
|HasAttr|key String|Boolean|Check whether an attribute exists in context
|GetAttr|key String|Any Type|Retrieve attibutes in context|

#### HTTPRequest Object
|method|args|return|remark|
| - | - | - | - |
|Header|name String|String|Get Request Header|
|RemoveHeader|name String||Remove Request Header|
|SetHeader|name String, value String||Set Request Header|
|Cookie|name String|String|Get Request Cookie|
|RemoveCookie|name String||Remove Request Cookie|
|SetCookie|name String, value String||Set Request Cookie|
|Query|name String|String|Get Request URL Arguments|
|Body||String|Get Request Body|
|SetBody|String||Set Request Body|

#### HTTPResponse Object
|method|args|return|remark|
| - | - | - | - |
|Delegate||Go fasthttp Reponse|一般用在插件的pre方法中使用指定Response的时候|
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

### 内建模块
由于js执行引擎并不兼容nodejs module，所以不能使用nodes module，所以gateway提供了一些内建的模块帮助编写插件。`由于这些内建库都是GO实现的，所以所有的方法名称首字母都是大写`，这里的习惯和js有点违背，但无关大雅

#### JSON
JSON内建插件JSON的编解码

|method|args|return|remark|
| - | - | - | - |
|Stringify|JSON|String||
|Parse|String|JSON||

#### LOG
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

#### HTTP
提供插件HTTP的能力，并提供同步的编程模式，返回结果就是HTTP的响应

|method|args|return|remark|
| - | - | - | - |
|NewHTTPResponse||HTTPResponse||
|Get|String|HTTPResult||
|Post|url String, body string, header JSON|HTTPResult|需要自己在header中设置Content-Type|
|PostJSON|url String, body string, header JSON|HTTPResult|会自动在header中设置Content-Type的类型为application/json|
|Put|url String, body string, header JSON|HTTPResult|需要自己在header中设置Content-Type|
|PutJSON|url String, body string, header JSON|HTTPResult|会自动在header中设置Content-Type的类型为application/json|
|Delete|url String, body string, header JSON|HTTPResult|需要自己在header中设置Content-Type|
|DeleteJSON|url String, body string, header JSON|HTTPResult|会自动在header中设置Content-Type的类型为application/json|

##### HTTPResult
|method|args|return|remark|
| - | - | - | - |
|HasError||Boolean|判断本次HTTP请求有没有错误|
|Error||String|返回本次HTTP请求错误|
|StatusCode||Number|返回本次请求的响应状态码|
|Header||Object|返回响应的Header|
|Cookie||Object Array|返回响应的Cookie|
|Body||String|响应Body体|

#### Redis
提供js访问Redis的能力

|method|args|return|remark|
| - | - | - | - |
|RedisModule|Object|RedisOp|创建Redis操作对象，`{"maxActive":"最大链接数，int", "maxIdle":"链接最大idle链接数，int", "idleTimeout":"idle超时时间，单位秒,int", "addr":"redis地址"}`|

#### RedisOp

|method|args|return|remark|
| - | - | - | - |
|Do|cmd String, args ...Any Type|CmdResp|执行redis命令|

#### CmdResp

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

## 插件方法
### pre
gateway在转发请求到后端server之前会调用插件的`pre`方法，方法返回一个JSON结构，有`code`和`error`字段。一旦返回的`error`字段不为空，gateway会使用返回的`code`的字段返回客户端。例如插件检测请求没有鉴权可以返回`{"code": 403, "error": "not login"}`。正常情况可以返回`{"code": 200}`。

在`pre`方法中，插件可以从调用上下文中获取`原始请求`和`转发请求`来处理。

一些其他功能:
- 插件可以使用`BreakFilterChainCode`来停止后续的插件`pre`方法的运行。例如返回：`{"code": BreakFilterChainCode}`
- 插件可以在插件上下文中设置属性`UsingResponse`，来让gateway使用指定的`Response`来返回客户端。提示：可以使用内建的`http`模块的`NewHTTPResponse`方法创建一个HTTP响应。例如在`pre`方法中加入如下逻辑：
```javascript
{
    "pre": function(ctx) {
        var HTTP = require("http")
        var resp = HTTP.NewHTTPResponse()
        resp.SetStatusCode(200)
        resp.SetBody('{"name":"zhangsan"}')
        resp.SetHeader("Content-Type", "application/json")
        ctx.SetAttr(UsingResponse, resp.Delegate())

        return {
            "code": BreakFilterChainCode
        }
    }
}
```

### post
gateway会在收到后端server的响应后调用插件的`post`方法，方法返回一个JSON结构，有`code`和`error`字段。一旦返回的`error`字段不为空，gateway会使用返回的`code`的字段返回客户端。正常情况可以返回`{"code": 200}`。

在`post`方法中，插件可以从调用上下文中获取`原始请求`，`转发请求`和`响应`来处理。

### postErr
Gateway calls `postErr` when redirecting to backend server fails.

In `postErr`, plugin can retrieve `OriginalRequest`，`RedirectRequest` from context.
