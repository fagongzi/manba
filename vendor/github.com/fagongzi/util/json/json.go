package json

import (
	"encoding/json"
	"log"
	"runtime"
)

// InitModel init model
type InitModel interface {
	Init() error
}

// MustMarshal marshal
func MustMarshal(value interface{}) []byte {
	v, err := json.Marshal(value)
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("json marshal failed, value=<%v> errors:\n %+v \n %s",
			value,
			err,
			buf)
	}
	return v
}

// MustUnmarshal unmarshal
func MustUnmarshal(value interface{}, data []byte) {
	err := Unmarshal(value, data)
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("json unmarshal failed, data=<%v> errors:\n %+v \n %s",
			data,
			err,
			buf)
	}

	if init, ok := value.(InitModel); ok {
		init.Init()
	}
}

// Unmarshal unmarshal
func Unmarshal(value interface{}, data []byte) error {
	err := json.Unmarshal(data, value)
	if err != nil {
		return err
	}

	if init, ok := value.(InitModel); ok {
		init.Init()
	}

	return nil
}

// Clone deep clone
func Clone(dest interface{}, src interface{}) {
	MustUnmarshal(dest, MustMarshal(src))
}
