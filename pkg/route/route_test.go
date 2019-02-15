package route

import (
	"bytes"
	"testing"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

func TestAdd(t *testing.T) {
	r := NewRoute()
	err := r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/users",
	})
	if err != nil {
		t.Errorf("add error")
	}
	if len(r.root.children) != 1 {
		t.Errorf("expect 1 children but %d", len(r.root.children))
	}

	err = r.Add(&metapb.API{
		ID:         2,
		URLPattern: "/accounts",
	})
	if err != nil {
		t.Errorf("add error with 2 const")
	}
	if len(r.root.children) != 2 {
		t.Errorf("expect 2 children but %d, %+v", len(r.root.children), r.root)
	}

	err = r.Add(&metapb.API{
		ID:         3,
		URLPattern: "/(string):name",
	})
	if err != nil {
		t.Errorf("add error with const and string")
	}
	if len(r.root.children) != 3 {
		t.Errorf("expect 3 children but %d, %+v", len(r.root.children), r.root)
	}

	err = r.Add(&metapb.API{
		ID:         4,
		URLPattern: "/(number):age",
	})
	if err != nil {
		t.Errorf("add error with const, string, number")
	}
	if len(r.root.children) != 4 {
		t.Errorf("expect 4 children but %d, %+v", len(r.root.children), r.root)
	}

	err = r.Add(&metapb.API{
		ID:         5,
		URLPattern: "/(enum:off|on):action",
	})
	if err != nil {
		t.Errorf("add error with const, string, number, enum")
	}
	if len(r.root.children) != 5 {
		t.Errorf("expect 4 children but %d, %+v", len(r.root.children), r.root)
	}
}

func TestFind(t *testing.T) {
	r := NewRoute()
	r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/",
	})
	r.Add(&metapb.API{
		ID:         2,
		URLPattern: "/check",
	})
	r.Add(&metapb.API{
		ID:         3,
		URLPattern: "/(string):name",
	})
	r.Add(&metapb.API{
		ID:         4,
		URLPattern: "/(number):age",
	})
	r.Add(&metapb.API{
		ID:         5,
		URLPattern: "/(enum:on|off):action",
	})

	params := make(map[string][]byte, 0)
	paramsFunc := func(name, value []byte) {
		params[string(name)] = value
	}

	id, _ := r.Find([]byte("/"), paramsFunc)
	if id != 1 {
		t.Errorf("expect matched 1, but %d", id)
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/check"), paramsFunc)
	if id != 2 {
		t.Errorf("expect matched 2, but %d", id)
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/check2"), paramsFunc)
	if id != 3 {
		t.Errorf("expect matched 3, but %d", id)
	}
	if bytes.Compare(params["name"], []byte("check2")) != 0 {
		t.Errorf("expect params check2, but %s", params["name"])
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/123"), paramsFunc)
	if id != 4 {
		t.Errorf("expect matched 4, but %d", id)
	}
	if bytes.Compare(params["age"], []byte("123")) != 0 {
		t.Errorf("expect params 123, but %s", params["age"])
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/on"), paramsFunc)
	if id != 5 {
		t.Errorf("expect matched 5, but %d", id)
	}
	if bytes.Compare(params["action"], []byte("on")) != 0 {
		t.Errorf("expect params on, but %s", params["action"])
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/off"), paramsFunc)
	if id != 5 {
		t.Errorf("expect matched 5, but %d", id)
	}
	if bytes.Compare(params["action"], []byte("off")) != 0 {
		t.Errorf("expect params off, but %s", params["action"])
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/on/notmatches"), paramsFunc)
	if id != 0 {
		t.Errorf("expect not matched , but %d", id)
	}
}
