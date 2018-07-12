package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/util/hack"
	"github.com/garyburd/redigo/redis"
	"github.com/valyala/fasthttp"
)

const (
	checkMethodTokenInRedis  string = "token_in_redis"
	checkMethodExpireByRaw   string = "expire_by_raw"
	checkMethodExpireByRedis string = "expire_by_redis"
	ctxRenewTokenAttr        string = "__jwt_renew_token__"
	jwtClaimsFieldExp        string = "exp"
)

var (
	errJWTMissing = errors.New("missing jwt token")
	errJWTInvalid = errors.New("invalid jwt token")
)

type fetcher func(jwt.MapClaims, filter.Context)
type tokenGetter func(filter.Context) (string, error)
type checker func(map[string]interface{}, string, jwt.MapClaims, filter.Context) (bool, error)

// JWTCfg cfg
type JWTCfg struct {
	Secret               string    `json:"secret"`
	Method               string    `json:"method"`
	TokenLookup          string    `json:"tokenLookup"`
	AuthSchema           string    `json:"authSchema"`
	RenewTokenHeaderName string    `json:"renewTokenHeaderName,omitempty"`
	Fetch                *Fetch    `json:"fetch,omitempty"`
	Redis                *Redis    `json:"redis,omitempty"`
	Checkers             []Checker `json:"checkers,omitempty"`
}

// Checker checker
type Checker struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// Fetch fetch fields int Claims
type Fetch struct {
	To     string   `json:"to"`
	Prefix string   `json:"prefix"`
	Fields []string `json:"fields"`
}

// Redis redis
type Redis struct {
	Addr        string `json:"addr"`
	MaxActive   int    `json:"maxActive"`
	MaxIdle     int    `json:"maxIdle"`
	IdleTimeout int    `json:"idleTimeout"`
}

// JWTFilter filter
type JWTFilter struct {
	filter.BaseFilter

	cfg              *JWTCfg
	secretBytes      []byte
	getter           tokenGetter
	redisPool        *redis.Pool
	leaseTTLDuration time.Duration
	signing          *jwt.SigningMethodHMAC
	checkers         []checker
	checkerArgs      []map[string]interface{}
	fetcher          fetcher
}

func newJWTFilter(file string) (filter.Filter, error) {
	f := &JWTFilter{}

	err := f.parseCfg(file)
	if err != nil {
		return nil, err
	}

	err = f.initSigningMethod()
	if err != nil {
		return nil, err
	}

	err = f.initCheckers()
	if err != nil {
		return nil, err
	}

	err = f.initFetcher()
	if err != nil {
		return nil, err
	}

	f.initRedisPool()
	f.initTokenLookup()
	return f, nil
}

// Name name
func (f *JWTFilter) Name() string {
	return FilterJWT
}

// Pre execute before proxy
func (f *JWTFilter) Pre(c filter.Context) (statusCode int, err error) {
	if strings.ToUpper(c.API().AuthFilter) != f.Name() {
		return f.BaseFilter.Pre(c)
	}

	token, err := f.getter(c)
	if err != nil {
		return fasthttp.StatusForbidden, err
	}

	claims, err := f.parseJWTToken(token)
	if err != nil {
		return fasthttp.StatusForbidden, err
	}

	for idx, ck := range f.checkers {
		ok, err := ck(f.checkerArgs[idx], token, claims, c)
		if err != nil {
			return fasthttp.StatusInternalServerError, err
		}

		if !ok {
			return fasthttp.StatusForbidden, nil
		}
	}

	if f.cfg.Fetch != nil {
		for _, field := range f.cfg.Fetch.Fields {
			c.ForwardRequest().Header.Add(fmt.Sprintf("%s%s", f.cfg.Fetch.Prefix, field), fmt.Sprintf("%v", claims[field]))
		}
	}

	return f.BaseFilter.Pre(c)
}

// Post execute after proxy
func (f *JWTFilter) Post(c filter.Context) (statusCode int, err error) {
	if value := c.GetAttr(ctxRenewTokenAttr); value != nil {
		c.Response().Header.Add(f.cfg.RenewTokenHeaderName, value.(string))
	}

	return f.BaseFilter.Post(c)
}

func (f *JWTFilter) parseCfg(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	cfg := &JWTCfg{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	f.cfg = cfg
	f.secretBytes = []byte(f.cfg.Secret)
	return nil
}

func (f *JWTFilter) initRedisPool() {
	if f.cfg.Redis != nil {
		f.redisPool = &redis.Pool{
			MaxActive:   f.cfg.Redis.MaxActive,
			MaxIdle:     f.cfg.Redis.MaxIdle,
			IdleTimeout: time.Second * time.Duration(f.cfg.Redis.IdleTimeout),
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp",
					f.cfg.Redis.Addr,
					redis.DialWriteTimeout(time.Second*10))
			},
		}
	}
}

func (f *JWTFilter) initSigningMethod() error {
	if f.cfg.Method == "HS256" {
		f.signing = jwt.SigningMethodHS256
	} else if f.cfg.Method == "HS384" {
		f.signing = jwt.SigningMethodHS384
	} else if f.cfg.Method == "HS512" {
		f.signing = jwt.SigningMethodHS512
	} else {
		return fmt.Errorf("unsupport method: %s", f.cfg.Method)
	}

	return nil
}

func (f *JWTFilter) initTokenLookup() {
	parts := strings.Split(f.cfg.TokenLookup, ":")
	f.getter = jwtFromHeader(parts[1], f.cfg.AuthSchema)
	switch parts[0] {
	case "query":
		f.getter = jwtFromQuery(parts[1])
	case "cookie":
		f.getter = jwtFromCookie(parts[1])
	}
}

func (f *JWTFilter) initFetcher() error {
	if f.cfg.Fetch == nil {
		return nil
	}

	switch f.cfg.Fetch.To {
	case "header":
		f.fetcher = f.fetchToHeader
		return nil
	case "cookie":
		f.fetcher = f.fetchToCookie
		return nil
	default:
		return fmt.Errorf("not supoprt fetch: %s", f.cfg.Fetch.To)
	}
}

func (f *JWTFilter) initCheckers() error {
	for _, c := range f.cfg.Checkers {
		f.checkerArgs = append(f.checkerArgs, c.Params)

		switch c.Method {
		case checkMethodTokenInRedis:
			f.checkers = append(f.checkers, f.tokenInRedisChecker)
		case checkMethodExpireByRaw:
			f.checkers = append(f.checkers, f.expireByRawChecker)
		case checkMethodExpireByRedis:
			f.checkers = append(f.checkers, f.expireByRedisChecker)
		default:
			return fmt.Errorf("not support check method: %s", c.Method)
		}
	}

	return nil
}

func (f *JWTFilter) parseJWTToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return f.secretBytes, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("error jwt token")
}

func (f *JWTFilter) renewToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(f.signing, claims)
	return token.SignedString(f.secretBytes)
}

func (f *JWTFilter) getRedis() redis.Conn {
	return f.redisPool.Get()
}

func (f *JWTFilter) expireByRawChecker(args map[string]interface{}, token string, claims jwt.MapClaims, c filter.Context) (bool, error) {
	if value, ok := claims[jwtClaimsFieldExp]; ok {
		now := time.Now()
		exp := int64(value.(float64))
		if exp > now.Unix() {
			return false, nil
		}

		claims[jwtClaimsFieldExp] = now.Add(time.Second * time.Duration(args["ttl"].(float64))).Unix()
		newToken, err := f.renewToken(claims)
		if err != nil {
			return false, err
		}

		c.SetAttr(ctxRenewTokenAttr, newToken)
		return true, nil
	}

	return true, nil
}

func (f *JWTFilter) expireByRedisChecker(args map[string]interface{}, token string, claims jwt.MapClaims, c filter.Context) (bool, error) {
	if f.cfg.Redis == nil {
		return false, fmt.Errorf("redis not setting")
	}

	var buf bytes.Buffer
	buf.WriteString(args["prefix"].(string))
	buf.WriteString(token)
	key := hack.SliceToString(buf.Bytes())

	conn := f.getRedis()
	value, err := redis.Int(conn.Do("TTL", key))
	if err != nil {
		conn.Close()
		return false, err
	}

	// key not exists or ttl is 0
	if value == -2 || value == 0 {
		conn.Close()
		return false, nil
	}

	// no ttl
	if value == -1 {
		conn.Close()
		return true, nil
	}

	_, err = conn.Do("SETEX", key, int(args["ttl"].(float64)), token)
	if err != nil {
		conn.Close()
		return false, err
	}

	conn.Close()
	return true, nil
}

func (f *JWTFilter) tokenInRedisChecker(args map[string]interface{}, token string, claims jwt.MapClaims, c filter.Context) (bool, error) {
	if f.cfg.Redis == nil {
		return false, fmt.Errorf("redis not setting")
	}

	var buf bytes.Buffer
	buf.WriteString(args["prefix"].(string))
	buf.WriteString(token)
	key := hack.SliceToString(buf.Bytes())

	conn := f.getRedis()
	value, err := redis.Bool(conn.Do("EXISTS", key))
	conn.Close()
	return value, err
}

func (f *JWTFilter) fetchToHeader(claims jwt.MapClaims, c filter.Context) {
	var buf bytes.Buffer
	for _, field := range f.cfg.Fetch.Fields {
		buf.WriteString(f.cfg.Fetch.Prefix)
		buf.WriteString(field)
		c.ForwardRequest().Header.Add(buf.String(), fmt.Sprintf("%v", claims[field]))
		buf.Reset()
	}
}

func (f *JWTFilter) fetchToCookie(claims jwt.MapClaims, c filter.Context) {
	var buf bytes.Buffer
	for _, field := range f.cfg.Fetch.Fields {
		buf.WriteString(f.cfg.Fetch.Prefix)
		buf.WriteString(field)
		c.ForwardRequest().Header.SetCookie(buf.String(), fmt.Sprintf("%v", claims[field]))
	}
}

func jwtFromQuery(param string) tokenGetter {
	return func(c filter.Context) (string, error) {
		token := string(c.OriginRequest().Request.URI().QueryArgs().Peek(param))
		if token == "" {
			return "", errJWTMissing
		}
		return token, nil
	}
}

func jwtFromCookie(name string) tokenGetter {
	return func(c filter.Context) (string, error) {
		value := string(c.OriginRequest().Request.Header.Cookie(name))
		if len(value) == 0 {
			return "", errJWTMissing
		}
		return value, nil
	}
}

func jwtFromHeader(header string, authScheme string) tokenGetter {
	return func(c filter.Context) (string, error) {
		auth := string(c.OriginRequest().Request.Header.Peek(header))
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", errJWTMissing
	}
}
