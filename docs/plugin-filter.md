Filter plugin
--------------
Many features of gateway are based on Filters , the user's most of the functional requirements can be used to solve the Filter. So Filter is designed as Plugin mechanism, with the help of Go1.8 plugin mechanism, can be a good extension Gateway.

### Request processing flow
Request -> filter preprocessing -> forward request -> filter post -> response to the client

The entire logic process conforms to the following rules:

* Filter preprocessing returns an error, the process terminates immediately, and uses the filter to return the status code to respond to the client
* Filter post processing error, use the filter to return the status code to respond to the client
* Forward request, back-end return to the status code `> = 500`, call the filter error handling interface

### Filter interface definition
`` `Golang
// Filter filter interface
Type Filter interface {
Name () string

Pre (c Context) (statusCode int, err error)
Post (c Context) (statusCode int, err error)
PostErr (c Context)
}

// Context filter context
Type Context interface {
SetStartAt (startAt int64)
SetEndAt (endAt int64)
GetStartAt () int64
GetEndAt () int64

GetProxyServerAddr () string
GetProxyOuterRequest () * fasthttp.Request
GetProxyResponse () * fasthttp.Response
NeedMerge () bool

GetOriginRequestCtx () * fasthttp.RequestCtx

GetMaxQPS () int

ValidateProxyOuterRequest () bool

InBlacklist (ip string) bool
InWhitelist (ip string) bool

IsCircuitOpen () bool
IsCircuitHalf () bool

GetOpenToCloseFailureRate () int
GetHalfTrafficRate () int
GetHalfToOpenSucceedRate () int
GetOpenToCloseCollectSeconds () int

ChangeCircuitStatusToClose ()
ChangeCircuitStatusToOpen ()

RecordMetricsForRequest ()
RecordMetricsForResponse ()
RecordMetricsForFailure ()
RecordMetricsForReject ()

GetRecentlyRequestSuccessedCount (sec int) int
GetRecentlyRequestCount (sec int) int
GetRecentlyRequestFailureCount (sec int) int
}

// BaseFilter base filter support default implemention
Type BaseFilter struct {}

// Pre execute before proxy
Func (f BaseFilter) Pre (c Context) (statusCode int, err err) {
Return http.StatusOK, nil
}

// Post execute after proxy
Func (f BaseFilter) Post (c Context) (statusCode int, err error) {
Return http.StatusOK, nil
}

// PostErr execute proxy has errors
Func (f BaseFilter) PostErr (c Context) {

}
`` ``

These related definitions are in the `github.com / fagongzi / gateway / pkg / filter` package, and each filter needs to be imported. One of the context of the Context Context provides the ability to interact with the Filter and Gateway; `BaseFilter` defines the default behavior.

### Gateway loads the Filter plugin mechanism
`` `Golang
Func newExternalFilter (filterSpec * conf.FilterSpec) (filter.Filter, error) {
P, err: = plugin.Open (filterSpec.ExternalPluginFile)
If err! = Nil {
Return nil, err
}

S, err := p.Lookup ("NewExternalFilter")
If err! = Nil {
Return nil, err
}

Sf: = s. (Func () (filter.Filter, error))
Return sf ()
}
`` ``

Each of the external Filter plugins, supplied with `NewExternalFilter`, returns a` filter.Filter` implementation, or an error.

### Go1.8 Plugin problem
When writing a custom plugin, there is a problem related to Go1.8 [bugs] (https://github.com/golang/go/issues/19233). So the custom plugin must be compiled under the `Gateway project`.