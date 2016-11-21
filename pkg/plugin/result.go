package plugin

const (
	// CodeSuccess code success
	CodeSuccess = iota
	// CodeError code error
	CodeError
)

// Result result
type Result struct {
	Code  int         `json:"code, omitempty"`
	Error string      `json:"error"`
	Value interface{} `json:"value"`
}
