package plugin

import (
	"encoding/json"
)

// JSONModule json builtin
type JSONModule struct {
}

// Stringify returns json string
func (j *JSONModule) Stringify(value interface{}) string {
	v, _ := json.Marshal(value)
	return string(v)
}

// Parse parse a string to json
func (j *JSONModule) Parse(value string) map[string]interface{} {
	obj := make(map[string]interface{})
	json.Unmarshal([]byte(value), &obj)
	return obj
}
