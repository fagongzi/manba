package model

import (
	"testing"
)

func TestAPIValidate(t *testing.T) {
	a := newTestNormalAPI()
	err := a.Validate()
	if err != nil {
		t.Error("api validation failed")
		return
	}

	a.Name = ""
	err = a.Validate()
	if err == nil {
		t.Error("api validation failed")
		return
	}

	a = newTestNormalAPI()
	a.URL = ""
	err = a.Validate()
	if err == nil {
		t.Error("api validation failed")
		return
	}

	a = newTestNormalAPI()
	a.Method = ""
	err = a.Validate()
	if err == nil {
		t.Error("api validation failed")
		return
	}
	a.Method = "a"
	err = a.Validate()
	if err == nil {
		t.Error("api validation failed")
		return
	}
}

func newTestNormalAPI() *API {
	return &API{
		Name:          "test-api",
		URL:           "/api/*",
		Method:        "GET",
		Status:        Down,
		AccessControl: newTestAccessControl(),
		Mock:          newTestMock(),
		Nodes: []*Node{&Node{
			ClusterName: "",
		}},
	}
}

func newTestAccessControl() *AccessControl {
	return &AccessControl{
		Whitelist: []string{"127.0.0.1"},
		Blacklist: []string{"127.0.0.1"},
	}
}

func newTestMock() *Mock {
	return &Mock{
		Value:       "value",
		ContentType: "application/json",
	}
}
