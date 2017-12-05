package json

import (
	"encoding/json"
	"fmt"
)

// InitModel init model
type InitModel interface {
	Init() error
}

// MustMarshal marshal
func MustMarshal(value interface{}) []byte {
	v, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("marash failed: %+v", err))
	}
	return v
}

// MustUnmarshal unmarshal
func MustUnmarshal(value interface{}, data []byte) {
	err := Unmarshal(value, data)
	if err != nil {
		panic(fmt.Sprintf("unmarash failed: %+v", err))
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
