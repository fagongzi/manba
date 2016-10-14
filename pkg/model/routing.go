package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/brettlangdon/forge"
	"github.com/fagongzi/goetty"
	"github.com/valyala/fasthttp"
)

var (
	// GlobalCfgDesc global desc cfg
	GlobalCfgDesc = "desc"
	// GlobalCfgOrder global order cfg
	GlobalCfgOrder = "order"
	// GlobalCfgDeadline global deadline cfg
	GlobalCfgDeadline = "deadline"
	// GlobalCfgRule global rule cfg
	GlobalCfgRule = "rule"
	// GlobalCfgOr global or cfg
	GlobalCfgOr = "or"
)

// rule: left [==,>,<=,>=,in,~] right
// can use var: $header_, $cookie_, $query_
var (
	// ValueHeader value of header prefix
	ValueHeader = "$HEADER"
	// ValueCookie value of cookie prefix
	ValueCookie = "$COOKIE"
	// ValueQuery value of query string prefix
	ValueQuery = "$QUERY"
)

var (
	// EQ ==
	EQ = "=="
	// LT <
	LT = "<"
	// LE <=
	LE = "<="
	// GT >
	GT = ">"
	// GE >=
	GE = ">="
	// IN in
	IN = "in"
	// MATCH reg matches
	MATCH = "~"

	// PATTERN reg pattern for expression
	PATTERN = regexp.MustCompile(fmt.Sprintf("^(?U)(.+)(%s|%s|%s|%s|%s|%s|%s)(.+)$", EQ, LE, LT, GE, GT, IN, MATCH))
)

var (
	// ErrSyntax Syntax error
	ErrSyntax = errors.New("Syntax error")
)

// RoutingItem routing item
type RoutingItem struct {
	rule string

	attrName       string
	sourceValueFun func(req *fasthttp.Request) string
	opFun          func(srvValue string) bool
	targetValue    string
}

// Routing routing
type Routing struct {
	ClusterName string `json:"clusterName, omitempty"`
	ID          string `json:"id,omitempty"`
	Cfg         string `json:"cfg,omitempty"`
	URL         string `json:"url,omitempty"`

	desc     string
	deadline int64
	regexp   *regexp.Regexp

	andItems []*RoutingItem
	orItems  []*RoutingItem
}

// UnMarshalRouting unmarshal
func UnMarshalRouting(data []byte) *Routing {
	v := &Routing{}
	json.Unmarshal(data, v)

	v.init()

	return v
}

// UnMarshalRoutingFromReader unmarshal
func UnMarshalRoutingFromReader(r io.Reader) (*Routing, error) {
	v := &Routing{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	if v.ID == "" {
		v.ID = goetty.NewV4UUID()
	}

	return v, err
}

// Marshal marshal
func (r *Routing) Marshal() []byte {
	v, _ := json.Marshal(r)
	return v
}

// Check check config
func (r *Routing) Check() error {
	return r.init()
}

// NewRouting create a new Routing
func NewRouting(cfgString string, clusterName string, url string) (*Routing, error) {
	r := &Routing{}
	r.Cfg = cfgString
	r.ClusterName = clusterName
	r.URL = url
	r.ID = goetty.NewV4UUID()

	return r, r.init()
}

func (r *Routing) init() error {
	reg, err := regexp.Compile(r.URL)

	if nil != err {
		return err
	}

	r.regexp = reg

	cfg, err := forge.ParseString(r.Cfg)

	if nil != err {
		return err
	}

	desc, err := cfg.GetString(GlobalCfgDesc)
	if nil != err {
		return err
	}
	r.desc = desc

	deadline, err := cfg.GetInteger(GlobalCfgDeadline)
	if nil != err {
		return err
	}
	r.deadline = deadline

	andRules, err := cfg.GetList(GlobalCfgRule)
	if nil != err {
		return err
	}
	items, err := parseRoutingItems(andRules)
	if nil != err {
		return err
	}
	r.andItems = items

	if cfg.Exists(GlobalCfgOr) {
		orRules, err := cfg.GetList(GlobalCfgOr)
		if nil != err {
			return err
		}
		items, err := parseRoutingItems(orRules)
		if nil != err {
			return err
		}
		r.orItems = items
	}

	return nil
}

// Matches return true if req matches
func (r *Routing) Matches(req *fasthttp.Request) bool {
	if !r.regexp.MatchString(string(req.URI().Path())) {
		return false
	}

	for _, item := range r.andItems {
		if !item.matches(req) {
			return r.matchesOr(req)
		}
	}

	return true
}

func (r *Routing) matchesOr(req *fasthttp.Request) bool {
	if nil != r.orItems {
		for _, item := range r.orItems {
			if item.matches(req) {
				return true
			}
		}

		return false
	}

	return false
}

func parseRoutingItems(rules *forge.List) ([]*RoutingItem, error) {
	items := make([]*RoutingItem, rules.Length())
	for i := 0; i < rules.Length(); i++ {
		rule, err := rules.GetString(i)
		if nil != err {
			return nil, err
		}
		item, err := newRoutingItem(rule)
		if nil != err {
			return nil, err
		}

		items[i] = item
	}

	return items, nil
}

func newRoutingItem(rule string) (*RoutingItem, error) {
	item := &RoutingItem{rule: rule}
	err := item.parse()
	return item, err
}

func (r *RoutingItem) parse() error {
	mg := PATTERN.FindStringSubmatch(r.rule)
	if len(mg) != 4 {
		return ErrSyntax
	}

	infos := mg[1:]
	op := strings.TrimSpace(infos[1])
	switch op {
	case EQ:
		r.opFun = r.eq
		break
	case LT:
		r.opFun = r.lt
		break
	case LE:
		r.opFun = r.le
		break
	case GT:
		r.opFun = r.gt
		break
	case GE:
		r.opFun = r.ge
		break
	case IN:
		r.opFun = r.in
		break
	case MATCH:
		r.opFun = r.reg
		break
	default:
		return ErrSyntax
	}

	attr := strings.TrimSpace(infos[0])
	attrInfos := strings.SplitN(attr, "_", 2)
	if len(attrInfos) != 2 {
		return ErrSyntax
	}
	prefix := strings.ToUpper(strings.TrimSpace(attrInfos[0]))
	switch prefix {
	case ValueHeader:
		r.sourceValueFun = r.getHeaderValue
		break
	case ValueCookie:
		r.sourceValueFun = r.getCookieValue
		break
	case ValueQuery:
		r.sourceValueFun = r.getQueryValue
	default:
		return ErrSyntax
	}

	r.attrName = strings.TrimSpace(attrInfos[1])
	r.targetValue = strings.TrimSpace(infos[2])

	return nil
}

func (r *RoutingItem) matches(req *fasthttp.Request) bool {
	return r.opFun(r.sourceValueFun(req))
}

func (r *RoutingItem) getCookieValue(req *fasthttp.Request) string {
	return string(req.Header.Cookie(r.attrName))
}

func (r *RoutingItem) getHeaderValue(req *fasthttp.Request) string {
	return string(req.Header.Peek(r.attrName))
}

func (r *RoutingItem) getQueryValue(req *fasthttp.Request) string {
	v, _ := url.QueryUnescape(string(req.URI().QueryArgs().Peek(r.attrName)))
	return v
}

func (r *RoutingItem) eq(srvValue string) bool {
	return srvValue == r.targetValue
}

func (r *RoutingItem) lt(srvValue string) bool {
	s, err := strconv.Atoi(srvValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(r.targetValue)
	if err != nil {
		return false
	}

	return s < t
}

func (r *RoutingItem) le(srvValue string) bool {
	s, err := strconv.Atoi(srvValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(r.targetValue)
	if err != nil {
		return false
	}

	return s <= t
}

func (r *RoutingItem) gt(srvValue string) bool {
	s, err := strconv.Atoi(srvValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(r.targetValue)
	if err != nil {
		return false
	}

	return s > t
}

func (r *RoutingItem) ge(srvValue string) bool {
	s, err := strconv.Atoi(srvValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(r.targetValue)
	if err != nil {
		return false
	}

	return s >= t
}

func (r *RoutingItem) in(srvValue string) bool {
	return strings.Index(srvValue, r.targetValue) != -1
}

func (r *RoutingItem) reg(srvValue string) bool {
	matches, err := regexp.MatchString(r.targetValue, srvValue)
	return err == nil && matches
}
