package model

import (
	"testing"

	"github.com/fagongzi/gateway/pkg/lb"
)

func TestClusterValidate(t *testing.T) {
	value := &Cluster{
		Name:   "c1",
		LbName: lb.ROUNDROBIN,
	}

	err := value.Validate()
	if err != nil {
		t.Errorf("validate cluster failed")
		return
	}

	value.Name = ""
	err = value.Validate()
	if err == nil {
		t.Errorf("validate cluster failed")
		return
	}

	value.Name = "c1"
	value.LbName = ""
	err = value.Validate()
	if err == nil {
		t.Errorf("validate cluster failed")
		return
	}

	value.LbName = "1"
	err = value.Validate()
	if err == nil {
		t.Errorf("validate cluster failed")
		return
	}
}
