Filter plugin
--------------
Gateway中的很多功能都是使用Filter来实现的，用户的大部分功能需求都可以使用Filter来解决。所以Filter被设计成Plugin机制，借助于Go1.8的plugin机制，可以很好的扩展Gateway。

### Request处理流程
request -> filter预处理 -> 转发请求 -> filter后置处理 -> 响应客户端

整个逻辑处理符合以下规则:

* filter预处理返回错误，流程立即终止，并且使用filter返回的状态码响应客户端
* filter后置处理返回错误，使用filter返回的状态码响应客户端
* 转发请求，后端返回的状态码`>=500`，调用filter的错误处理接口

### Filter接口定义
```golang
// Filter filter interface
type Filter interface {
	Name() string

	Pre(c Context) (statusCode int, err error)
	Post(c Context) (statusCode int, err error)
	PostErr(c Context)
}

// Context filter context
type Context interface {
	SetStartAt(startAt int64)
	SetEndAt(endAt int64)
	GetStartAt() int64
	GetEndAt() int64

	GetProxyServerAddr() string
	GetProxyOuterRequest() *fasthttp.Request
	GetProxyResponse() *fasthttp.Response
	NeedMerge() bool

	GetOriginRequestCtx() *fasthttp.RequestCtx

	GetMaxQPS() int

	ValidateProxyOuterRequest() bool

	InBlacklist(ip string) bool
	InWhitelist(ip string) bool

	IsCircuitOpen() bool
	IsCircuitHalf() bool

	GetOpenToCloseFailureRate() int
	GetHalfTrafficRate() int
	GetHalfToOpenSucceedRate() int
	GetOpenToCloseCollectSeconds() int

	ChangeCircuitStatusToClose()
	ChangeCircuitStatusToOpen()

	RecordMetricsForRequest()
	RecordMetricsForResponse()
	RecordMetricsForFailure()
	RecordMetricsForReject()

	GetRecentlyRequestSuccessedCount(sec int) int
	GetRecentlyRequestCount(sec int) int
	GetRecentlyRequestFailureCount(sec int) int
}

// BaseFilter base filter support default implemention
type BaseFilter struct{}

// Pre execute before proxy
func (f BaseFilter) Pre(c Context) (statusCode int, err error) {
	return http.StatusOK, nil
}

// Post execute after proxy
func (f BaseFilter) Post(c Context) (statusCode int, err error) {
	return http.StatusOK, nil
}

// PostErr execute proxy has errors
func (f BaseFilter) PostErr(c Context) {

}
```

这些相关的定义都在`github.com/fagongzi/gateway/pkg/filter`包中，每一个Filter都需要导入。其中的`Context`的上下文接口，提供了Filter和Gateway交互的能力;`BaseFilter`定义了默认行为。

### Gateway加载Filter插件机制
```golang
func newExternalFilter(filterSpec *conf.FilterSpec) (filter.Filter, error) {
	p, err := plugin.Open(filterSpec.ExternalPluginFile)
	if err != nil {
		return nil, err
	}

	s, err := p.Lookup("NewExternalFilter")
	if err != nil {
		return nil, err
	}

	sf := s.(func() (filter.Filter, error))
	return sf()
}
```

每一个外部的Filter插件，对外提供`NewExternalFilter`，返回一个`filter.Filter`实现，或者错误。

### Go1.8 Plugin的问题
当编写的自定义插件的时候，有一个问题涉及到Go1.8的一个[Bug](https://github.com/golang/go/issues/19233)。所以编写的自定义插件必须在`Gateway的Project`下编译的插件才能被正确加载。

### 配置一个外部Filter
配置Proxy启动的配置文件
