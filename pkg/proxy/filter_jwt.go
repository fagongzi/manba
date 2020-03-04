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
	// besides checking token is legitimate or not, it checks whether token exists in redis
	actionTokenInRedis string = "token_in_redis"
	// update token's TTL
	actionRenewByRaw string = "renew_by_raw"
	// update token's TTL and in the same time put new token in redis, previous token invalid
	actionRenewByRedis string = "renew_by_redis"
	// fetch fields from token and put them in header which is redirected to a backend server who is unbeknownst to JWT
	actionFetchToHeader string = "fetch_to_header"
	actionFetchToCookie string = "fetch_to_cookie"
	ctxRenewTokenAttr   string = "__jwt_renew_token__"
	jwtClaimsFieldExp   string = "exp"
)

var (
	errJWTMissing = errors.New("missing jwt token")
	errJWTInvalid = errors.New("invalid jwt token")
)

type tokenGetter func(filter.Context) (string, error)
type action func(map[string]interface{}, string, jwt.MapClaims, filter.Context) (bool, error)

// JWTCfg cfg
type JWTCfg struct {
	Secret               string   `json:"secret"`
	Method               string   `json:"method"`
	TokenLookup          string   `json:"tokenLookup"`
	AuthSchema           string   `json:"authSchema"`
	RenewTokenHeaderName string   `json:"renewTokenHeaderName,omitempty"`
	Redis                *Redis   `json:"redis,omitempty"`
	Actions              []Action `json:"actions,omitempty"`
}

// Action action
type Action struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
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
	actions          []action
	actionArgs       []map[string]interface{}
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

	err = f.initActions()
	if err != nil {
		return nil, err
	}

	f.initRedisPool()
	err = f.initTokenLookup()
	if err != nil {
		return nil, err
	}
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

	for idx, act := range f.actions {
		ok, err := act(f.actionArgs[idx], token, claims, c)
		if err != nil {
			return fasthttp.StatusInternalServerError, err
		}

		if !ok {
			return fasthttp.StatusForbidden, nil
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
		return fmt.Errorf("unsupported method: %s", f.cfg.Method)
	}

	return nil
}

func (f *JWTFilter) initTokenLookup() error {
	parts := strings.Split(f.cfg.TokenLookup, ":")
	if len(parts) < 2 {
		return fmt.Errorf("TokenLookup should contain : ")
	}
	f.getter = jwtFromHeader(parts[1], f.cfg.AuthSchema)
	switch parts[0] {
	case "query":
		f.getter = jwtFromQuery(parts[1])
	case "cookie":
		f.getter = jwtFromCookie(parts[1])
	}
	return nil
}

func (f *JWTFilter) initActions() error {
	for _, c := range f.cfg.Actions {
		f.actionArgs = append(f.actionArgs, c.Params)

		switch c.Method {
		case actionTokenInRedis:
			f.actions = append(f.actions, f.tokenInRedisAction)
		case actionRenewByRaw:
			f.actions = append(f.actions, f.renewByRawAction)
		case actionRenewByRedis:
			f.actions = append(f.actions, f.renewByRedisAction)
		case actionFetchToHeader:
			f.actions = append(f.actions, f.fetchToHeader)
		case actionFetchToCookie:
			f.actions = append(f.actions, f.fetchToCookie)
		default:
			return fmt.Errorf("not support action method: %s", c.Method)
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

func (f *JWTFilter) renewByRawAction(args map[string]interface{}, token string, claims jwt.MapClaims, c filter.Context) (bool, error) {
	if _, ok := claims[jwtClaimsFieldExp]; ok {
		claims[jwtClaimsFieldExp] = time.Now().Add(time.Second * time.Duration(args["ttl"].(float64))).Unix()
		newToken, err := f.renewToken(claims)
		if err != nil {
			return false, err
		}
		c.SetAttr(ctxRenewTokenAttr, newToken)
		return true, nil
	}

	return true, nil
}

func (f *JWTFilter) renewByRedisAction(args map[string]interface{}, token string, claims jwt.MapClaims, c filter.Context) (bool, error) {
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

func (f *JWTFilter) tokenInRedisAction(args map[string]interface{}, token string, claims jwt.MapClaims, c filter.Context) (bool, error) {
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

func (f *JWTFilter) fetchToHeader(args map[string]interface{}, token string, claims jwt.MapClaims, c filter.Context) (bool, error) {
	var buf bytes.Buffer
	prefix := args["prefix"].(string)
	for _, fd := range args["fields"].([]interface{}) {
		field := fd.(string)
		buf.WriteString(prefix)
		buf.WriteString(field)
		c.ForwardRequest().Header.Add(buf.String(), fmt.Sprintf("%v", claims[field]))
		buf.Reset()
	}

	return true, nil
}

func (f *JWTFilter) fetchToCookie(args map[string]interface{}, token string, claims jwt.MapClaims, c filter.Context) (bool, error) {
	var buf bytes.Buffer
	prefix := args["prefix"].(string)
	for _, fd := range args["fields"].([]interface{}) {
		field := fd.(string)
		buf.WriteString(prefix)
		buf.WriteString(field)
		c.ForwardRequest().Header.SetCookie(buf.String(), fmt.Sprintf("%v", claims[field]))
		buf.Reset()
	}

	return true, nil
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
