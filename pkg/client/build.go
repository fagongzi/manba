package client

import (
	"strings"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

// ServerBuildOption build option
type ServerBuildOption func(*metapb.Server)

// APIBuildOption build option
type APIBuildOption func(*metapb.API)

// DispatchNodeBuildOption build option
type DispatchNodeBuildOption func(*metapb.DispatchNode)

// BuildServer returns a new server
func BuildServer(addr string, opts ...ServerBuildOption) metapb.Server {
	value := metapb.Server{
		Addr: addr,
	}
	for _, opt := range opts {
		opt(&value)
	}

	return value
}

// BuildAPI returns a new api
func BuildAPI(name string, status metapb.Status, opts ...APIBuildOption) metapb.API {
	value := metapb.API{
		Name:   name,
		Status: status,
	}
	for _, opt := range opts {
		opt(&value)
	}

	return value
}

// WithHTTPBackend returns a server build option
func WithHTTPBackend() ServerBuildOption {
	return func(value *metapb.Server) {
		value.Protocol = metapb.HTTP
	}
}

// WithCheckHTTPCode returns a server build option
func WithCheckHTTPCode(path string, interval time.Duration, timeout time.Duration) ServerBuildOption {
	return func(value *metapb.Server) {
		value.HeathCheck = &metapb.HeathCheck{
			Path:          path,
			CheckInterval: int64(interval),
			Timeout:       int64(timeout),
		}
	}
}

// WithCheckHTTPBody returns a server build option
func WithCheckHTTPBody(path, body string, interval time.Duration, timeout time.Duration) ServerBuildOption {
	return func(value *metapb.Server) {
		value.HeathCheck = &metapb.HeathCheck{
			Path:          path,
			Body:          body,
			CheckInterval: int64(interval),
			Timeout:       int64(timeout),
		}
	}
}

// WithQPS returns a server build option
func WithQPS(max int64) ServerBuildOption {
	return func(value *metapb.Server) {
		value.MaxQPS = max
	}
}

// CircuitBreakerCondition circuit breaker condition
type CircuitBreakerCondition func(*metapb.CircuitBreaker)

// CircuitBreakerCloseCondition returns a condition with close status
func CircuitBreakerCloseCondition(timeout time.Duration) CircuitBreakerCondition {
	return func(cb *metapb.CircuitBreaker) {
		cb.CloseTimeout = int64(timeout)
	}
}

// CircuitBreakerHalfCondition returns a condition with half status
func CircuitBreakerHalfCondition(trafficRate int, failureRateToClose, succeedRateToOpen int) CircuitBreakerCondition {
	return func(cb *metapb.CircuitBreaker) {
		cb.HalfTrafficRate = int32(trafficRate)
		cb.SucceedRateToOpen = int32(succeedRateToOpen)
		cb.FailureRateToClose = int32(failureRateToClose)
	}
}

// WithCircuitBreaker returns a server build option
func WithCircuitBreaker(checkPeriod time.Duration, conds ...CircuitBreakerCondition) ServerBuildOption {
	return func(value *metapb.Server) {
		value.CircuitBreaker = &metapb.CircuitBreaker{
			RateCheckPeriod: int64(checkPeriod),
		}

		for _, cond := range conds {
			cond(value.CircuitBreaker)
		}
	}
}

// WithMatchMethod returns a api build option
func WithMatchMethod(method string) APIBuildOption {
	return func(value *metapb.API) {
		value.Method = strings.ToUpper(method)
	}
}

// WithMatchAllMethod returns a api build option
func WithMatchAllMethod() APIBuildOption {
	return func(value *metapb.API) {
		value.Method = "*"
	}
}

// WithMatchURL returns a api build option
func WithMatchURL(pattern string) APIBuildOption {
	return func(value *metapb.API) {
		value.URLPattern = pattern
	}
}

// WithMatchDomain returns a api build option
func WithMatchDomain(domain string) APIBuildOption {
	return func(value *metapb.API) {
		value.Domain = domain
	}
}

// WithDefaultResult returns a api build option
func WithDefaultResult(body []byte, headers []*metapb.PairValue, cookies []*metapb.PairValue) APIBuildOption {
	return func(value *metapb.API) {
		value.DefaultValue = &metapb.HTTPResult{
			Body:    body,
			Headers: headers,
			Cookies: cookies,
		}
	}
}

// WithWhiteIPList returns a api build option
func WithWhiteIPList(ips ...string) APIBuildOption {
	return func(value *metapb.API) {
		if value.IPAccessControl == nil {
			value.IPAccessControl = &metapb.IPAccessControl{}
		}

		value.IPAccessControl.Whitelist = append(value.IPAccessControl.Whitelist, ips...)
	}
}

// WithBlackIPList returns a api build option
func WithBlackIPList(ips ...string) APIBuildOption {
	return func(value *metapb.API) {
		if value.IPAccessControl == nil {
			value.IPAccessControl = &metapb.IPAccessControl{}
		}

		value.IPAccessControl.Blacklist = append(value.IPAccessControl.Blacklist, ips...)
	}
}

// WithAddDispatchNode returns a add dispatcher node build option.
func WithAddDispatchNode(targetCluster uint64, opts ...DispatchNodeBuildOption) APIBuildOption {
	return func(value *metapb.API) {
		node := &metapb.DispatchNode{
			ClusterID: targetCluster,
		}

		for _, opt := range opts {
			opt(node)
		}

		value.Nodes = append(value.Nodes, node)
	}
}

// WithMultiDispatch returns a multi node build option
// If the api has more dispatch nodes, every node has a result,
// every result is a attribute of the full json result.
func WithMultiDispatch(attrName string) DispatchNodeBuildOption {
	return func(value *metapb.DispatchNode) {
		value.AttrName = attrName
	}
}

// WithURLWriteDispatch returns a url write build option
func WithURLWriteDispatch(urlRewrite string) DispatchNodeBuildOption {
	return func(value *metapb.DispatchNode) {
		value.URLRewrite = urlRewrite
	}
}

// WithAddValidation returns a add validation build option
func WithAddValidation(param metapb.Parameter, required bool, rules ...metapb.ValidationRule) DispatchNodeBuildOption {
	return func(value *metapb.DispatchNode) {
		value.Validations = append(value.Validations, &metapb.Validation{
			Parameter: param,
			Required:  required,
			Rules:     rules,
		})
	}
}

// NewRegexpRule returns a regexp rule
func NewRegexpRule(pattern string) *metapb.ValidationRule {
	return &metapb.ValidationRule{
		RuleType:   metapb.RuleRegexp,
		Expression: pattern,
	}
}
