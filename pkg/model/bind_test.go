package model

import (
	"testing"
)

func TestBindValidate(t *testing.T) {
	value := &Bind{
		ServerID:  "1",
		ClusterID: "2",
	}

	err := value.Validate()
	if err != nil {
		t.Errorf("validate bind failed")
		return
	}

	value.ServerID = ""
	err = value.Validate()
	if err == nil {
		t.Errorf("validate bind failed")
		return
	}

	value.ServerID = "1"
	value.ClusterID = ""
	err = value.Validate()
	if err == nil {
		t.Errorf("validate bind failed")
		return
	}
}
