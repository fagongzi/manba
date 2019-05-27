Filter plugin
--------------
Most features of Gateway are implemented through Filter. Most of users' functional requirements are implemented through Filter. Filter is thus implemented as a plugin thanks to Go 1.8's plugin mechanism to scale Gateway well.

# Handling Procedures of Request
request -> filter preprocess -> redirect request -> filter postprocess -> respond client

All the logic processing follows the rules below:

* When filter preprocessing returns error, the procedure aborts immediately and uses the returned status code of filter to respond to client.
* When filter postprocessing returns error, the returned status code of filter is used to respond to client.
* When the status code of the response of redirected requests is `>=500`, filter's error handling API is called.

# Filter API Definition
```golang
// Filter filter interface
type Filter interface {
	Name() string
	Init(cfg string) error

	Pre(c Context) (statusCode int, err error)
	Post(c Context) (statusCode int, err error)
	PostErr(c Context)
}

// Context filter context
type Context interface {
	StartAt() time.Time
	EndAt() time.Time

	OriginRequest() *fasthttp.RequestCtx
	ForwardRequest() *fasthttp.Request
	Response() *fasthttp.Response

	API() *metapb.API
	DispatchNode() *metapb.DispatchNode
	Server() *metapb.Server
	Analysis() *util.Analysis

	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}
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

Relevant definitions are in `github.com/fagongzi/gateway/pkg/filter`. Each filter needs to be imported. `Context` API provides the ability of interactions between Filter and Gateway. `BaseFilter` defines the default.

# The Mechanism of Gateway Loading Filter Plugin
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

Every external Filter plugin exposes `NewExternalFilter`. It returns a `filter.Filter` or an error.

# A Problem of Go1.8 Plugin
When writing customized plugin, there is a problem concerning a Go 1.8 [bug](https://github.com/golang/go/issues/19233). It can be avoided by compiling the customized plugin under `Gateway/Project` in order to make the plugin load correctly.

# For Go 1.9.2 and Above
Independent directories of plugins are supported. However, it can not have its own vender directory, otherwise, the same problem arises as with Go 1.8

# A Customized Plugin Example
You can refer to this example to make your own plugin
[JWT Plugin Reference](https://github.com/fagongzi/jwt-plugin)

# Start A Customized Plugin Example
`Proxy` component has a `--filter` option to designate plugins used by Gateway and its order. By default, the order of built-in plugins of Gateway：`--filter WHITELIST --filter WHITELIST --filter ANALYSIS --filter RATE-LIMITING --filter CIRCUIT-BREAKER --filter HTTP-ACCESS --filter HEADER --filter XFORWARD --filter VALIDATION`. For instance, suppose we have a JWT plugin ready and it is compiled as a jwt.so file, options to start can be `--filter WHITELIST --filter WHITELIST --filter ANALYSIS --filter RATE-LIMITING --filter CIRCUIT-BREAKER --filter HTTP-ACCESS --filter HEADER --filter XFORWARD --filter VALIDATION --filter JWT:/plugins/jwt.so:/plugins/jwt.json`. The format of a customized plugin is `Name:Plugin File:Plugin Configuration`