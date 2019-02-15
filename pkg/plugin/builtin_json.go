package plugin

import (
	"encoding/json"
)

// JSON json builtin
type JSON struct {
}

// Stringify returns json string
func (j *JSON) Stringify(value interface{}) string {
	v, _ := json.Marshal(value)
	return string(v)
}

// Parse parse a string to json
func (j *JSON) Parse(value string) map[string]interface{} {
	obj := make(map[string]interface{})
	json.Unmarshal([]byte(value), obj)
	return obj
}
