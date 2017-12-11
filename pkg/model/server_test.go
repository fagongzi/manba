package model

import (
	"testing"
)

func TestServerValidate(t *testing.T) {
	value := &Server{
		Addr:   "1",
		MaxQPS: 1,
		Schema: "http",
	}

	err := value.Validate()
	if err != nil {
		t.Errorf("validate server failed")
		return
	}

	value.MaxQPS = 0
	err = value.Validate()
	if err == nil {
		t.Errorf("validate server failed")
		return
	}

	value.MaxQPS = 1
	value.Addr = ""
	err = value.Validate()
	if err == nil {
		t.Errorf("validate server failed")
		return
	}

	value.Addr = "1"
	value.Schema = ""
	err = value.Validate()
	if err == nil {
		t.Errorf("validate server failed")
		return
	}
}
