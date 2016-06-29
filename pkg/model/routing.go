package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brettlangdon/forge"
	"github.com/fagongzi/goetty"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	CFG_GLOBAL_DESC     = "desc"
	CFG_GLOBAL_ORDER    = "order"
	CFG_GLOBAL_DEADLINE = "deadline"
	CFG_GLOBAL_RULES    = "rule"
	CFG_GLOBAL_OR       = "or"
)

// rule: left [==,>,<=,>=,in,~] right
// can use var: $header_, $cookie_, $query_
var (
	VALUE_HEADER = "$HEADER"
	VALUE_COOKIE = "$COOKIE"
	VALUE_QUERY  = "$QUERY"
)

var (
	EQ      = "=="
	LT      = "<"
	LE      = "<="
	GT      = ">"
	GE      = ">="
	IN      = "in"
	MATCH   = "~"
	PATTERN = regexp.MustCompile(fmt.Sprintf("^(?U)(.+)(%s|%s|%s|%s|%s|%s|%s)(.+)$", EQ, LE, LT, GE, GT, IN, MATCH))
)

var (
	ERR_SYNTAX = errors.New("Syntax error")
)

type RoutingItem struct {
	rule string

	attrName       string
	sourceValueFun func(req *http.Request) string
	opFun          func(srvValue string) bool
	targetValue    string
}

type Routing struct {
	ClusterName string `json:"clusterName, omitempty"`
	Id          string `json:"id,omitempty"`
	Cfg         string `json:"cfg,omitempty"`
	Url         string `json:"url,omitempty"`

	desc     string
	deadline int64
	regexp   *regexp.Regexp

	andItems []*RoutingItem
	orItems  []*RoutingItem
}

func UnMarshalRouting(data []byte) *Routing {
	v := &Routing{}
	json.Unmarshal(data, v)

	v.init()

	return v
}

func UnMarshalRoutingFromReader(r io.Reader) (*Routing, error) {
	v := &Routing{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	if v.Id == "" {
		v.Id = goetty.NewV4UUID()
	}

	return v, err
}

func (self *Routing) Marshal() []byte {
	v, _ := json.Marshal(self)
	return v
}

func (self *Routing) Check() error {
	return self.init()
}

func NewRouting(cfgString string, clusterName string, url string) (*Routing, error) {
	r := &Routing{}
	r.Cfg = cfgString
	r.ClusterName = clusterName
	r.Url = url
	r.Id = goetty.NewV4UUID()

	return r, r.init()
}

func (r *Routing) init() error {
	reg, err := regexp.Compile(r.Url)

	if nil != err {
		return err
	}

	r.regexp = reg

	cfg, err := forge.ParseString(r.Cfg)

	if nil != err {
		return err
	}

	desc, err := cfg.GetString(CFG_GLOBAL_DESC)
	if nil != err {
		return err
	}
	r.desc = desc

	deadline, err := cfg.GetInteger(CFG_GLOBAL_DEADLINE)
	if nil != err {
		return err
	}
	r.deadline = deadline

	andRules, err := cfg.GetList(CFG_GLOBAL_RULES)
	if nil != err {
		return err
	}
	items, err := parseRoutingItems(andRules)
	if nil != err {
		return err
	}
	r.andItems = items

	if cfg.Exists(CFG_GLOBAL_OR) {
		orRules, err := cfg.GetList(CFG_GLOBAL_OR)
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

func (self *Routing) Matches(req *http.Request) bool {
	if !self.regexp.MatchString(req.URL.Path) {
		return false
	}

	for _, r := range self.andItems {
		if !r.matches(req) {
			return self.matchesOr(req)
		}
	}

	return true
}

func (self *Routing) matchesOr(req *http.Request) bool {
	if nil != self.orItems {
		for _, r := range self.orItems {
			if r.matches(req) {
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

func (self *RoutingItem) parse() error {
	mg := PATTERN.FindStringSubmatch(self.rule)
	if len(mg) != 4 {
		return ERR_SYNTAX
	}

	infos := mg[1:]
	op := strings.TrimSpace(infos[1])
	switch op {
	case EQ:
		self.opFun = self.eq
		break
	case LT:
		self.opFun = self.lt
		break
	case LE:
		self.opFun = self.le
		break
	case GT:
		self.opFun = self.gt
		break
	case GE:
		self.opFun = self.ge
		break
	case IN:
		self.opFun = self.in
		break
	case MATCH:
		self.opFun = self.reg
		break
	default:
		return ERR_SYNTAX
	}

	attr := strings.TrimSpace(infos[0])
	attrInfos := strings.SplitN(attr, "_", 2)
	if len(attrInfos) != 2 {
		return ERR_SYNTAX
	}
	prefix := strings.ToUpper(strings.TrimSpace(attrInfos[0]))
	switch prefix {
	case VALUE_HEADER:
		self.sourceValueFun = self.getHeaderValue
		break
	case VALUE_COOKIE:
		self.sourceValueFun = self.getCookieValue
		break
	case VALUE_QUERY:
		self.sourceValueFun = self.getQueryValue
	default:
		return ERR_SYNTAX
	}

	self.attrName = strings.TrimSpace(attrInfos[1])
	self.targetValue = strings.TrimSpace(infos[2])

	return nil
}

func (self *RoutingItem) matches(req *http.Request) bool {
	return self.opFun(self.sourceValueFun(req))
}

func (self *RoutingItem) getCookieValue(req *http.Request) string {
	value, err := req.Cookie(self.attrName)

	if err != nil {
		return ""
	}

	return value.Value
}

func (self *RoutingItem) getHeaderValue(req *http.Request) string {
	return req.Header.Get(self.attrName)
}

func (self *RoutingItem) getQueryValue(req *http.Request) string {
	v, _ := url.QueryUnescape(req.URL.Query().Get(self.attrName))
	return v
}

func (self *RoutingItem) eq(srvValue string) bool {
	return srvValue == self.targetValue
}

func (self *RoutingItem) lt(srvValue string) bool {
	s, err := strconv.Atoi(srvValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(self.targetValue)
	if err != nil {
		return false
	}

	return s < t
}

func (self *RoutingItem) le(srvValue string) bool {
	s, err := strconv.Atoi(srvValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(self.targetValue)
	if err != nil {
		return false
	}

	return s <= t
}

func (self *RoutingItem) gt(srvValue string) bool {
	s, err := strconv.Atoi(srvValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(self.targetValue)
	if err != nil {
		return false
	}

	return s > t
}

func (self *RoutingItem) ge(srvValue string) bool {
	s, err := strconv.Atoi(srvValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(self.targetValue)
	if err != nil {
		return false
	}

	return s >= t
}

func (self *RoutingItem) in(srvValue string) bool {
	return strings.Index(srvValue, self.targetValue) != -1
}

func (self *RoutingItem) reg(srvValue string) bool {
	matches, err := regexp.MatchString(self.targetValue, srvValue)
	return err == nil && matches
}
