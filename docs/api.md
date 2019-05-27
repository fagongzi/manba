API
-----------
API is the key concept of Gateway. We can manage APIs through API-Server of Gateway.

# API Attributes
## ID
API's ID, unique identifier

## Name
API's name

## URLPattern
URL Matching Expression. Gateway uses this to match original request URLs. URLPattern must be used with `Method`. These two need to be met for the request to be qualified as matching the API.

### URLPattern Expression
`/` is used to divide URL Path. Each part can be the following types：

* string constant: any legitimate URLs. May use `*` to match any string
* (number): argeName a number variable
* (string) a string variable
* (enum:enum1|enum1|enum1) a enum variable. Available enums are divided by |.
* If these variables are needed, for example, in URL Rewrite, you could use:
    * (number):id
    * (string):name
    * (enum:on|off):action

Some examples:
* `/api/v1/product/(number):id`
* `/api/v1/product/(number):id/(enum:online|offline):action`
* `/api/v1/*`

## Method (Optional)
It is used to match HTTP request's `HTTP Method` field, `*` matches all HTTP Method (GET, PUT, POST, DELETE).

## Domain (Optional)
It matches HOST field of HTTP requests.

## MatchRule
Matching Rules of `URLPattern`,`Method`,`Domain`
* MatchDefault `URLPattern` && (`Method` || `Domain`)
* MatchAll `URLPattern` && (`Method` && `Domain`)
* MatchAny `URLPattern` && (`Method` || `Domain`)

## Status
`UP`, `Down`. API valid only if `UP`.

## IPAccessControl (Optional)
White list and black list

## DefaultValue (Optional)
API's default return value. When there is no available server in the backend cluster, Gateway returns this value which consists of Code, HTTP Body, Header, and Cookie. It can be used as the default return value of Mock or backend services.

## Aggregation Requests
An original request is redirected to multiple backend servers and the reponses are merged into a JSON instance whose attributes correspond to each reponse. Multiple redirected requests can be sent simultaneously or in a batch (a request argument depends on prior one's response). `BatchIndex` sets the order of each redirected requests.

Example
* Original Request: `/api/v1/aggregation/1`
* Redirected Request: `/api/v1/users/1`，Response: `{"name":"zhangsan"}`, Attribute: `user`
* Redirected Request: `/api/v1/accounts/1`，Response: `{"type":"test", "accountId":"123"}`，Attribute: `account`
* Final Response：`{"user":{"name":"zhangsan"}, "account":{"type":"test", "accountId":"123"}}`

## Nodes
Requests are redirected to at least one backend cluster. One request can be sent to multiple backend clusters at the same time (Currently only HTTP GET requests are supported). Redirected requests supports the following features.

### URL Rewrite
It reconstructs real URLs redirected to backend servers.

#### URL Rewrite Expression
Variables are supported in expressions. In runtime, `$()` are replaced by real values. `.` is used to access a variable's attributes.

##### origin variable
origin variable is used to extract `query string`,`header`,`cookie`,`body` of original requests. For multiple arguments with the same name, individual fetch is not supported.

An Original Request
* URL: `/api/v1/users?id=1&name=fagongzi&page=1&pageSize=100`,
* Header: `x-test-token: token1, x-test-id: 100`
* Cookie: `x-cookie-token: token2, x-cookie-id: 200`
* Body: `{"type":1, "value":{"id":100, "name":"zhangsan"}}`

Variable Form
* $(origin.query) = `?id=1&name=fagongzi&page=1&pageSize=100`
* $(origin.path) = `/api/v1/users`
* $(origin.query.id) = `1`, $(origin.query.name) = `fagongzi`, $(origin.query.page) = `1`, $(origin.query.pageSize) = `100`
* $(origin.header.x-test-token) = `token1`, $(origin.header.x-test-id) = `100`
* $(origin.cookie.x-cookie-token) = `token2`, $(origin.cookie.x-cookie-id) = `200`
* $(origin.body.type) = `1`, $(origin.body.value.id) = `100`, $(origin.body.value.name) = `zhangsan`

#### param variable
param variable is used to extract variables from API's `URLPattern`.

Suppose:
* API's `URLPattern` is `/api/v1/users/(number):id/(enum:on|off):action`
* Request URL is `/api/v1/users/100/on`

Variable Form
* $(param.id) = `100`
* $(param.action) = `on`

#### depend variable
depend variable is used to extract data from return values of depend aggregation requests.

Suppose an aggregation request consists of two requests, user and account
* user: {"name":"zhangsan"}
* account: {"id":"123456"}

Variable Example
* $(depend.user.name) = `zhangsan`
* $(depend.account.id) = `123456`

#### URL Rewrite Expression Examples
* `/api/v1/users$(origin.query)`
* `$(origin.path)?name=$(origin.header.x-user-name)&id=$(origin.body.user.id)`
* `/api/v1/users?id=$(param.id)&action=$(param.action)`
* `/api/v1/accounts?id=$(depend.user.accountId)`

### Support for Check of Arguments in Original Requests
  Regular Expression Check Rule of Any Attribute Configuration in `querystring`, `json body`, `cookie`, `header` and `path value` is supported

###  Retry After Failure Is Supported
  `retryStrategy` can be set to retry based on HTTP response status. Maximum  number of trials and the interval of trial can be set.

### API Class Timeout
  `ReadTimeout` and `WriteTimeout` can be set to designate a request's read and write timeout. If not set, default global configuratio is used.

## Perms (Optional)
It is used to configure permission of an API. Users need to develop their own permission check plugins.

## AuthFilter (Optional)
Set an API's Auth plugin name. Reference to implementation of Auth plugin [JWT plugin](https://github.com/fagongzi/jwt-plugin)

## RenderTemplate
RenderTemplate can be used to redefine responses which include data format and fields.

## UseDefault (Optional)

When it is true and `DefaultValue`exists, `DefaultValue` is used as response value.

## MatchRule (Optional)

| MatchRule | Logic |
| - | - |
| MatchDefault | `Domain` \|\| (`URLPattern` && `Method`) |
| MatchAll | `Domain` && `URLPattern` && `Method` |
| MatchAny | `Domain` \|\| `URLPattern` \|\| `Method` |

## Position (Optional)

This value is used in increasing order in API matching phase. The smaller the value is, the higher the priority. Default value is 0.

## Tags (Optional)
For maintenance and search.

## WebSocketOptions (Optional)
websocket option, `websocket`. Attention: `websocket is still under testing phase. Closed by default. --websocket can be used to start`。When Gateway redirects websocket, `Origin` uses address of backend servers by default. If special value needs to be set, `Origin` argument can be designated.

## MaxQPS (Optional)
Maximal QPS API can support. Used to controll traffic. Gateway uses the Token Bucket Algorithm, restricting traffic by MaxQPS, thus protecting backend servers from overload. The priority of API is higher than what it is in `server`.

## CircuitBreaker（可选）
Backend API circuit break rule. It has three modes.

## CircuitBreaker (Optional)
Backend server circuit break status:

* Open

  Normal. All traffic in. When Gateway find the failed requests to all requests ratio reach a certain threshold, CircuitBreaker switches from Open to Close.

* Half

  Attempt to recover. Gateway tries to direct a certain percentage of traffic to the server and observe the result. If the expectation is met, CircuitBreaker switches to Open. If not, Close.

* Close

  Gateway does not direct any traffic to this backend server. When the time threshold is reached, Gateway automatically tries to recover by switching to Half.