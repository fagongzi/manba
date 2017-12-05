package model

import (
	"regexp"

	"github.com/valyala/fasthttp"
)

const (
	// Regexp reg type
	Regexp = iota
)

// ValidationRule validation rule
type ValidationRule struct {
	Type       int    `json:"type"`
	Expression string `json:"expression"`
	pattern    *regexp.Regexp
}

// Validation validate rule
type Validation struct {
	Attr     *Attr             `json:"attr"`
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

func (v Validation) getValue(req *fasthttp.Request) string {
	return v.Attr.Value(req)
}

func (v Validation) validate(req *fasthttp.Request) bool {
	if nil == v.Rules || len(v.Rules) == 0 {
		return true
	}

	value := v.getValue(req)

	if "" == value && v.Required {
		return false
	} else if "" == value && !v.Required {
		return true
	}

	for _, rule := range v.Rules {
		if !rule.validate([]byte(value)) {
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
