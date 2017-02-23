package model

import (
	"regexp"

	"github.com/valyala/fasthttp"
)

const (
	// FromQueryString query string
	FromQueryString = iota
	// FromForm form data
	FromForm
)

const (
	// Regexp reg type
	Regexp = iota
)

// ValidationRule validation rule
type ValidationRule struct {
	Type       int    `json:"type, omitempty"`
	Expression string `json:"expression, omitempty"`
	pattern    *regexp.Regexp
}

// Validation validate rule
type Validation struct {
	Attr     string            `json:"attr, omitempty"`
	GetFrom  int               `json:"getFrom, omitempty"`
	Required bool              `json:"required, omitempty"`
	Rules    []*ValidationRule `json:"rules, omitempty"`
}

func (v ValidationRule) validate(value []byte) bool {
	return v.pattern.Match(value)
}

// ParseValidation parse validation
func (v Validation) ParseValidation() {
	if v.Rules != nil {
		for _, rule := range v.Rules {
			rule.pattern = regexp.MustCompile(rule.Expression)
		}
	}
}

func (v Validation) getValue(req *fasthttp.Request) []byte {
	if v.GetFrom == FromQueryString {
		return req.URI().QueryArgs().Peek(v.Attr)
	} else if v.GetFrom == FromForm {
		return req.PostArgs().Peek(v.Attr)
	}

	return nil
}

func (v Validation) validate(req *fasthttp.Request) bool {
	if nil == v.Rules || len(v.Rules) == 0 {
		return true
	}

	value := v.getValue(req)

	if nil == value && v.Required {
		return false
	} else if nil == value && !v.Required {
		return true
	}

	for _, rule := range v.Rules {
		if !rule.validate(value) {
			return false
		}
	}

	return true
}

// Validate validate request
func (n *Node) Validate(req *fasthttp.Request) bool {
	if n.Validations == nil || len(n.Validations) == 0 {
		return true
	}

	for _, validation := range n.Validations {
		if !validation.validate(req) {
			return false
		}
	}

	return true
}
