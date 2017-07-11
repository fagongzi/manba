package api

const (
	// CodeSuccess success code
	CodeSuccess = 0
	// CodeError error code
	CodeError = 1
)

// Result is the return value of api server
type Result struct {
	Code  int         `json:"code"`
	Error string      `json:"error, omitempty"`
	Value interface{} `json:"value, omitempty"`
}
