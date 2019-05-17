package plugin

const (
	httpModuleName  = "http"
	jsonModuleName  = "json"
	logModuleName   = "log"
	redisModuleName = "redis"
)

var (
	httpModule = newHTTPModule()
	jsonModule = &JSONModule{}
	logModule  = &LogModule{}
)
