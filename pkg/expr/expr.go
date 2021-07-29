package expr

import (
	"bytes"
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/fagongzi/util/hack"
	"github.com/valyala/fasthttp"
)

const (
	eoi    byte = 0x1A
	dollar      = byte('$')
	lParen      = byte('(')
	rParen      = byte(')')
)

const (
	tokenEOF = iota
	tokenUnknown
	tokenDollar
	tokenLParen
	tokenRParen
)

var (
	origin = []byte("origin")
	path   = []byte("path")
	query  = []byte("query")
	cookie = []byte("cookie")
	header = []byte("header")
	body   = []byte("body")
	depend = []byte("depend")
	param  = []byte("param")

	dot = []byte{'.'}
)

// Ctx expr ctx
type Ctx struct {
	Origin *fasthttp.Request
	Depend []byte
	Params map[string][]byte
}

// Reset reset ctx
func (c *Ctx) Reset() {
	c.Origin = nil
	c.Depend = nil
	if c.Params != nil {
		for key := range c.Params {
			delete(c.Params, key)
		}
	}
}

// AddParam add param
func (c *Ctx) AddParam(name, value []byte) {
	c.Params[hack.SliceToString(name)] = value
}

// CopyParams copy params
func (c *Ctx) CopyParams() map[string][]byte {
	values := make(map[string][]byte)
	if c.Params != nil {
		for key, value := range c.Params {
			values[key] = value
		}
	}

	return values
}

// Expr expr
type Expr interface {
	Exec(buf *bytes.Buffer, ctx *Ctx)
	Name() string
}

// Exec returns expr result
func Exec(ctx *Ctx, exprs ...Expr) []byte {
	buf := bytes.NewBuffer(nil)
	for _, expr := range exprs {
		expr.Exec(buf, ctx)
	}

	return buf.Bytes()
}

// Parse parse expr
func Parse(value []byte) ([]Expr, error) {
	// symbol table
	// $(origin.path)/$(origin.query)/$(origin.query.xxx)/$(origin.cookie.xxx)/$(origin.header.xxx)/$(origin.body.xxx)
	// $(depend.xxx)
	// $(param.xxx)

	var exprs []Expr
	lexer := newScanner(value)
	prev := tokenUnknown

	for {
		lexer.NextToken()

		token := lexer.Token()
		switch token {
		case tokenDollar:
			switch prev {
			case tokenUnknown:
				break
			case tokenRParen:
				break
			default:
				return nil, fmt.Errorf("syntax error: expect symbol $ or )")
			}

			value := lexer.ScanString()
			if len(value) > 0 {
				exprs = append(exprs, &constExpr{value: value})
			}
			break
		case tokenLParen:
			if prev != tokenDollar {
				return nil, fmt.Errorf("syntax error: missing $")
			}

			lexer.ScanString()
			break
		case tokenRParen:
			if prev != tokenLParen {
				return nil, fmt.Errorf("syntax error: missing (")
			}
			exp, err := newExpr(lexer.ScanString())
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, exp)
			break
		case tokenEOF:
			switch prev {
			case tokenUnknown:
				break
			case tokenRParen:
				break
			default:
				return nil, fmt.Errorf("syntax error: maybe expect symbol )")
			}

			value := lexer.ScanString()
			if len(value) > 0 {
				exprs = append(exprs, &constExpr{value: value})
			}
			return exprs, nil
		}

		prev = token
	}
}

type parser struct {
	input []byte
	lexer scanner
}

func newParser(input []byte) *parser {
	return &parser{
		input: input,
		lexer: newScanner(input),
	}
}

type scanner struct {
	len   int
	input []byte

	token int
	bp    int
	sp    int
	ch    byte
}

func newScanner(input []byte) scanner {
	scan := scanner{
		len:   len(input),
		input: input,
		bp:    -1,
		sp:    0,
	}

	scan.Next()
	return scan
}

func (scan *scanner) Next() byte {
	scan.bp++

	if scan.bp < scan.len {
		scan.ch = scan.input[scan.bp]
	} else {
		scan.ch = eoi
	}

	return scan.ch
}

func (scan *scanner) NextToken() {
	for {
		switch scan.ch {
		case '$':
			scan.token = tokenDollar
			scan.Next()
			return
		case '(':
			scan.token = tokenLParen
			scan.Next()
			return
		case ')':
			scan.token = tokenRParen
			scan.Next()
			return
		case eoi:
			scan.token = tokenEOF
			scan.Next()
			return
		}

		scan.Next()
	}
}

func (scan *scanner) Current() byte {
	return scan.ch
}

func (scan *scanner) Token() int {
	return scan.token
}

func (scan *scanner) TokenIndex() int {
	return scan.bp - 1
}

func (scan *scanner) ScanString() []byte {
	value := scan.input[scan.sp : scan.bp-1]
	scan.sp = scan.bp
	return value
}

// $(origin.path)/$(origin.query)/$(origin.query.xxx)/$(origin.cookie.xxx)/$(origin.header.xxx)/$(origin.body.xxx)
// $(depend.xxx.xxx)
// $(param.xxx)
func newExpr(value []byte) (Expr, error) {
	values := bytes.Split(value, dot)
	if bytes.Equal(values[0], origin) {
		return newOriginExpr(values)
	} else if bytes.Equal(values[0], depend) {
		if len(values) < 2 {
			return nil, fmt.Errorf("syntax error: depend error")
		}

		return &dependParamExpr{param: toStringSlice(values[1:])}, nil
	} else if bytes.Equal(values[0], param) {
		if len(values) != 2 {
			return nil, fmt.Errorf("syntax error: param error")
		}

		return &pathParamExpr{param: hack.SliceToString(values[1])}, nil
	}

	return nil, fmt.Errorf("syntax error: not support source %s", values[0])
}

func newOriginExpr(values [][]byte) (Expr, error) {
	n := len(values)
	if n < 2 {
		return nil, fmt.Errorf("syntax error: origin error")
	}

	if bytes.Equal(values[1], query) {
		if n > 3 {
			return nil, fmt.Errorf("syntax error: origin query error")
		}

		if n == 2 {
			return &originQueryExpr{}, nil
		}

		return &originQueryParamExpr{param: hack.SliceToString(values[2])}, nil
	} else if bytes.Equal(values[1], path) {
		if n != 2 {
			return nil, fmt.Errorf("syntax error: origin path error")
		}

		return &originPathExpr{}, nil
	} else if bytes.Equal(values[1], cookie) {
		if n != 3 {
			return nil, fmt.Errorf("syntax error: origin cookie error")
		}

		return &originCookieParamExpr{param: hack.SliceToString(values[2])}, nil
	} else if bytes.Equal(values[1], header) {
		if n != 3 {
			return nil, fmt.Errorf("syntax error: origin header error")
		}

		return &originHeaderParamExpr{param: hack.SliceToString(values[2])}, nil
	} else if bytes.Equal(values[1], body) {
		if n < 3 {
			return nil, fmt.Errorf("syntax error: origin body error")
		}

		return &originBodyParamExpr{toStringSlice(values[2:])}, nil
	}

	return nil, fmt.Errorf("syntax error: not support origin %s", values[1])
}

type constExpr struct {
	value []byte
}

func (e *constExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	buf.Write(e.value)
}

func (e *constExpr) Name() string {
	return "const-expr"
}

type originQueryExpr struct {
}

func (e *originQueryExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	value := ctx.Origin.URI().QueryString()
	if len(value) > 0 {
		buf.WriteByte('?')
		buf.Write(ctx.Origin.URI().QueryString())
	}
}

func (e *originQueryExpr) Name() string {
	return "origin-query-expr"
}

type originQueryParamExpr struct {
	param string
}

func (e *originQueryParamExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	buf.Write(ctx.Origin.URI().QueryArgs().Peek(e.param))
}

func (e *originQueryParamExpr) Name() string {
	return "origin-query-param-expr"
}

type originPathExpr struct {
}

func (e *originPathExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	buf.Write(ctx.Origin.URI().Path())
}

func (e *originPathExpr) Name() string {
	return "origin-path-expr"
}

type originCookieParamExpr struct {
	param string
}

func (e *originCookieParamExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	buf.Write(ctx.Origin.Header.Cookie(e.param))
}

func (e *originCookieParamExpr) Name() string {
	return "origin-cookie-expr"
}

type originHeaderParamExpr struct {
	param string
}

func (e *originHeaderParamExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	buf.Write(ctx.Origin.Header.Peek(e.param))
}

func (e *originHeaderParamExpr) Name() string {
	return "origin-header-expr"
}

type originBodyParamExpr struct {
	param []string
}

func (e *originBodyParamExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	value, _, _, err := jsonparser.Get(ctx.Origin.Body(), e.param...)
	if err != nil {
		return
	}
	buf.Write(value)
}

func (e *originBodyParamExpr) Name() string {
	return "origin-body-expr"
}

type dependParamExpr struct {
	param []string
}

func (e *dependParamExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	if ctx.Depend != nil {
		value, _, _, err := jsonparser.Get(ctx.Depend, e.param...)
		if err != nil {
			return
		}
		buf.Write(value)
	}
}

func (e *dependParamExpr) Name() string {
	return "depend-expr"
}

type pathParamExpr struct {
	param string
}

func (e *pathParamExpr) Exec(buf *bytes.Buffer, ctx *Ctx) {
	if ctx.Params != nil {
		buf.Write(ctx.Params[e.param])
	}
}

func (e *pathParamExpr) Name() string {
	return "param-expr"
}

func toStringSlice(values [][]byte) []string {
	paths := make([]string, 0, len(values))
	for _, value := range values {
		paths = append(paths, hack.SliceToString(value))
	}
	return paths
}
