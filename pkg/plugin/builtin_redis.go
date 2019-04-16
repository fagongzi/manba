package plugin

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisModule redis module
type RedisModule struct {
	rt *Runtime
}

// CreateRedis create redis
func (m *RedisModule) CreateRedis(cfg map[string]interface{}) *RedisOp {
	p := &redis.Pool{
		MaxActive:   int(cfg["maxActive"].(int64)),
		MaxIdle:     int(cfg["maxIdle"].(int64)),
		IdleTimeout: time.Second * time.Duration(int(cfg["idleTimeout"].(int64))),
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				cfg["addr"].(string),
				redis.DialWriteTimeout(time.Second*10))
		},
	}

	m.rt.addCloser(p)

	conn := p.Get()
	_, err := conn.Do("PING")
	if err != nil {
		conn.Close()
		return &RedisOp{
			err: err,
		}
	}

	conn.Close()
	return &RedisOp{
		pool: p,
	}
}

// RedisOp redis
type RedisOp struct {
	err  error
	pool *redis.Pool
}

// Do do redis cmd
func (r *RedisOp) Do(cmd string, args ...interface{}) *CmdResp {
	if r.err != nil {
		return &CmdResp{
			err: r.err,
		}
	}

	conn := r.pool.Get()
	rsp, err := conn.Do(cmd, args...)
	if err != nil {
		conn.Close()
		return &CmdResp{
			err: err,
		}
	}

	conn.Close()
	return &CmdResp{
		rsp: rsp,
	}
}

// CmdResp redis cmd resp
type CmdResp struct {
	err error
	rsp interface{}
}

// HasError returns has error
func (r *CmdResp) HasError() bool {
	return r.err != nil
}

// Error returns  error
func (r *CmdResp) Error() string {
	if r.err != nil {
		return r.err.Error()
	}

	return ""
}

// StringValue returns string value
func (r *CmdResp) StringValue() string {
	if r.HasError() {
		return ""
	}

	value, _ := redis.String(r.rsp, nil)
	return value
}

// StringsValue returns strings value
func (r *CmdResp) StringsValue() []string {
	if r.HasError() {
		return nil
	}

	value, _ := redis.Strings(r.rsp, nil)
	return value
}

// StringMapValue returns string map value
func (r *CmdResp) StringMapValue() map[string]string {
	if r.HasError() {
		return make(map[string]string)
	}

	value, _ := redis.StringMap(r.rsp, nil)
	return value
}

// IntValue returns int value
func (r *CmdResp) IntValue() int {
	if r.HasError() {
		return 0
	}

	value, _ := redis.Int(r.rsp, nil)
	return value
}

// IntsValue returns ints value
func (r *CmdResp) IntsValue() []int {
	if r.HasError() {
		return nil
	}

	value, _ := redis.Ints(r.rsp, nil)
	return value
}

// IntMapValue returns int map value
func (r *CmdResp) IntMapValue() map[string]int {
	if r.HasError() {
		return make(map[string]int)
	}

	value, _ := redis.IntMap(r.rsp, nil)
	return value
}

// Int64Value returns int64 value
func (r *CmdResp) Int64Value() int64 {
	if r.HasError() {
		return 0
	}

	value, _ := redis.Int64(r.rsp, nil)
	return value
}

// Int64sValue returns int64s value
func (r *CmdResp) Int64sValue() []int64 {
	if r.HasError() {
		return nil
	}

	value, _ := redis.Int64s(r.rsp, nil)
	return value
}

// Int64MapValue returns int64 map value
func (r *CmdResp) Int64MapValue() map[string]int64 {
	if r.HasError() {
		return make(map[string]int64)
	}

	value, _ := redis.Int64Map(r.rsp, nil)
	return value
}
