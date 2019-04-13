package route

import (
	"bytes"
	"testing"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

func TestAddWithMethod(t *testing.T) {
	r := NewRoute()
	err := r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/users",
		Method:     "GET",
	})
	if err != nil {
		t.Errorf("add error")
	}
	if len(r.root.children) != 1 {
		t.Errorf("expect 1 children but %d", len(r.root.children))
	}

	err = r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/users",
		Method:     "PUT",
	})
	if err != nil {
		t.Errorf("add error")
	}
	if len(r.root.children) != 1 {
		t.Errorf("expect 1 children but %d", len(r.root.children))
	}

	err = r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/users",
		Method:     "DELETE",
	})
	if err != nil {
		t.Errorf("add error")
	}
	if len(r.root.children) != 1 {
		t.Errorf("expect 1 children but %d", len(r.root.children))
	}

	err = r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/users",
		Method:     "POST",
	})
	if err != nil {
		t.Errorf("add error")
	}
	if len(r.root.children) != 1 {
		t.Errorf("expect 1 children but %d", len(r.root.children))
	}

	err = r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/users",
		Method:     "*",
	})
	if err == nil {
		t.Errorf("add error, expect error")
	}

	err = r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/v1/users",
		Method:     "*",
	})
	if err != nil {
		t.Errorf("add error")
	}

	err = r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/v1/users",
		Method:     "GET",
	})
	if err == nil {
		t.Errorf("add error, expect error")
	}
}

func TestAdd(t *testing.T) {
	r := NewRoute()
	err := r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/users",
		Method:     "*",
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
		Method:     "*",
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
		Method:     "*",
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
		Method:     "*",
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
		Method:     "*",
	})
	if err != nil {
		t.Errorf("add error with const, string, number, enum")
	}
	if len(r.root.children) != 5 {
		t.Errorf("expect 4 children but %d, %+v", len(r.root.children), r.root)
	}
}

func TestFindWithStar(t *testing.T) {
	r := NewRoute()
	r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/*",
		Method:     "*",
	})

	r.Add(&metapb.API{
		ID:         2,
		URLPattern: "/users/1",
		Method:     "*",
	})

	id, _ := r.Find([]byte("/users/1"), "GET", nil)
	if id != 1 {
		t.Errorf("expect matched 1, but %d", id)
	}
}

func TestFind(t *testing.T) {
	r := NewRoute()
	r.Add(&metapb.API{
		ID:         1,
		URLPattern: "/",
		Method:     "*",
	})
	r.Add(&metapb.API{
		ID:         2,
		URLPattern: "/check",
		Method:     "*",
	})
	r.Add(&metapb.API{
		ID:         3,
		URLPattern: "/(string):name",
		Method:     "*",
	})
	r.Add(&metapb.API{
		ID:         4,
		URLPattern: "/(number):age",
		Method:     "*",
	})
	r.Add(&metapb.API{
		ID:         5,
		URLPattern: "/(enum:on|off):action",
		Method:     "*",
	})

	params := make(map[string][]byte, 0)
	paramsFunc := func(name, value []byte) {
		params[string(name)] = value
	}

	id, _ := r.Find([]byte("/"), "GET", paramsFunc)
	if id != 1 {
		t.Errorf("expect matched 1, but %d", id)
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/check"), "GET", paramsFunc)
	if id != 2 {
		t.Errorf("expect matched 2, but %d", id)
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/check2"), "GET", paramsFunc)
	if id != 3 {
		t.Errorf("expect matched 3, but %d", id)
	}
	if bytes.Compare(params["name"], []byte("check2")) != 0 {
		t.Errorf("expect params check2, but %s", params["name"])
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/123"), "GET", paramsFunc)
	if id != 4 {
		t.Errorf("expect matched 4, but %d", id)
	}
	if bytes.Compare(params["age"], []byte("123")) != 0 {
		t.Errorf("expect params 123, but %s", params["age"])
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/on"), "GET", paramsFunc)
	if id != 5 {
		t.Errorf("expect matched 5, but %d", id)
	}
	if bytes.Compare(params["action"], []byte("on")) != 0 {
		t.Errorf("expect params on, but %s", params["action"])
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/off"), "GET", paramsFunc)
	if id != 5 {
		t.Errorf("expect matched 5, but %d", id)
	}
	if bytes.Compare(params["action"], []byte("off")) != 0 {
		t.Errorf("expect params off, but %s", params["action"])
	}

	params = make(map[string][]byte, 0)
	id, _ = r.Find([]byte("/on/notmatches"), "GET", paramsFunc)
	if id != 0 {
		t.Errorf("expect not matched , but %d", id)
	}
}
