package model

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fagongzi/gateway/pkg/util"
	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/valyala/fasthttp"
)

// Op op like ==, > , < , ~ and so on
type Op int

const (
	// OpEQ ==
	OpEQ = Op(0)
	// OpLT <
	OpLT = Op(1)
	// OpLE <=
	OpLE = Op(2)
	// OpGT >
	OpGT = Op(3)
	// OpGE >=
	OpGE = Op(4)
	// OpIn in
	OpIn = Op(5)
	// OpMatch reg matches
	OpMatch = Op(6)
)

// Condition condition
type Condition struct {
	Attr  *Attr
	Op    Op
	Value string
}

// Match returns the http request is match this condition
func (c *Condition) Match(req *fasthttp.Request) bool {
	attrValue := c.Attr.Value(req)
	if attrValue == "" {
		return false
	}

	switch c.Op {
	case OpEQ:
		return c.eq(attrValue)
	case OpLT:
		return c.lt(attrValue)
	case OpLE:
		return c.le(attrValue)
	case OpGT:
		return c.gt(attrValue)
	case OpGE:
		return c.ge(attrValue)
	case OpIn:
		return c.in(attrValue)
	case OpMatch:
		return c.reg(attrValue)
	default:
		return false
	}
}

func (c *Condition) eq(attrValue string) bool {
	return attrValue == c.Value
}

func (c *Condition) lt(attrValue string) bool {
	s, err := strconv.Atoi(attrValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(c.Value)
	if err != nil {
		return false
	}

	return s < t
}

func (c *Condition) le(attrValue string) bool {
	s, err := strconv.Atoi(attrValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(c.Value)
	if err != nil {
		return false
	}

	return s <= t
}

func (c *Condition) gt(attrValue string) bool {
	s, err := strconv.Atoi(attrValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(c.Value)
	if err != nil {
		return false
	}

	return s > t
}

func (c *Condition) ge(attrValue string) bool {
	s, err := strconv.Atoi(attrValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(c.Value)
	if err != nil {
		return false
	}

	return s >= t
}

func (c *Condition) in(attrValue string) bool {
	return strings.Index(c.Value, attrValue) != -1
}

func (c *Condition) reg(attrValue string) bool {
	matches, err := regexp.MatchString(c.Value, attrValue)
	return err == nil && matches
}

// Routing routing
type Routing struct {
	ID         string       `json:"id, omitempty"`
	Conditions []*Condition `json:"conditions"`
	Rate       int          `json:"rate"`
	Cluster    string       `json:"cluster"`
	Status     Status       `json:"status"`

	rand *rand.Rand
}

// Validate validate the model
func (r *Routing) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Cluster, validation.Required))
}

// Matches return if the http request is matches this routing
func (r *Routing) Matches(req *fasthttp.Request) bool {
	for _, c := range r.Conditions {
		if !c.Match(req) {
			return false
		}
	}

	n := r.rand.Intn(100)
	return n < r.Rate
}

// Init init
func (r *Routing) Init() error {
	if r.ID == "" {
		r.ID = util.NewID()
	}

	r.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	return nil
}
