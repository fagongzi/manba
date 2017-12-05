package redis

import (
	"github.com/fagongzi/goetty"
)

type redisDecoder struct {
}

// NewRedisDecoder returns a redis protocol decoder
func NewRedisDecoder() goetty.Decoder {
	return &redisDecoder{}
}

// Decode decode
func (decoder *redisDecoder) Decode(in *goetty.ByteBuf) (bool, interface{}, error) {
	complete, cmd, err := ReadCommand(in)
	if err != nil {
		return true, nil, err
	}

	if !complete {
		return false, nil, nil
	}

	return true, cmd, nil
}

type redisReplyDecoder struct {
}

// NewRedisReplyDecoder returns a redis protocol cmd reply decoder
func NewRedisReplyDecoder() goetty.Decoder {
	return &redisReplyDecoder{}
}

// Decode decode
func (decoder *redisReplyDecoder) Decode(in *goetty.ByteBuf) (bool, interface{}, error) {
	complete, cmd, err := readCommandReply(in)
	if err != nil {
		return true, nil, err
	}

	if !complete {
		return false, nil, nil
	}

	return true, cmd, nil
}
