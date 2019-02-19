package plugin

var (
	builtins = make(map[string]interface{})
)

func init() {
	builtins["http"] = &HTTPModule{}
	builtins["json"] = &JSONModule{}
	builtins["log"] = &LogModule{}
	builtins["redis"] = &RedisModule{}
}

// Require require module
func Require(module string) interface{} {
	return builtins[module]
}
