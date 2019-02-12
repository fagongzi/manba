package route

import (
	"testing"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

func TestAdd(t *testing.T) {
	r := NewRoute()
	err := r.Add(metapb.API{
		ID:         1,
		URLPattern: "/users",
	})
	if err != nil {
		t.Errorf("add error")
	}
	if len(r.root.children) != 1 {
		t.Errorf("expect 1 children but %d", len(r.root.children))
	}
	if r.root.children[0].api != 1 {
		t.Errorf("expect api 1 children but %d", r.root.children[0].api)
	}

	err = r.Add(metapb.API{
		ID:         0,
		URLPattern: "/users",
	})
	if err == nil {
		t.Errorf("expect conflict error")
	}

	err = r.Add(metapb.API{
		ID:         2,
		URLPattern: "/accounts",
	})
	if err != nil {
		t.Errorf("add error with 2 const")
	}
	if len(r.root.children) != 2 {
		t.Errorf("expect 2 children but %d, %+v", len(r.root.children), r.root)
	}
	if r.root.children[1].api != 2 {
		t.Errorf("expect api 2 children but %d", r.root.children[1].api)
	}

	err = r.Add(metapb.API{
		ID:         3,
		URLPattern: "/(string):name",
	})
	if err != nil {
		t.Errorf("add error with const and string")
	}
	if len(r.root.children) != 3 {
		t.Errorf("expect 3 children but %d, %+v", len(r.root.children), r.root)
	}
	if r.root.children[2].api != 3 {
		t.Errorf("expect api 3 children but %d", r.root.children[2].api)
	}

	err = r.Add(metapb.API{
		ID:         4,
		URLPattern: "/(number):age",
	})
	if err != nil {
		t.Errorf("add error with const, string, number")
	}
	if len(r.root.children) != 4 {
		t.Errorf("expect 4 children but %d, %+v", len(r.root.children), r.root)
	}
	if r.root.children[3].api != 4 {
		t.Errorf("expect api 4 children but %d", r.root.children[3].api)
	}

	err = r.Add(metapb.API{
		ID:         5,
		URLPattern: "/(enum:off|on):action",
	})
	if err != nil {
		t.Errorf("add error with const, string, number, enum")
	}
	if len(r.root.children) != 5 {
		t.Errorf("expect 4 children but %d, %+v", len(r.root.children), r.root)
	}
	if r.root.children[4].api != 5 {
		t.Errorf("expect api 4 children but %d", r.root.children[4].api)
	}
}

func TestFind(t *testing.T) {
	r := NewRoute()
	r.Add(metapb.API{
		ID:         1,
		URLPattern: "/",
	})
	r.Add(metapb.API{
		ID:         2,
		URLPattern: "/check",
	})
	r.Add(metapb.API{
		ID:         3,
		URLPattern: "/(string)",
	})
	r.Add(metapb.API{
		ID:         4,
		URLPattern: "/(number)",
	})
	r.Add(metapb.API{
		ID:         5,
		URLPattern: "/(enum:on|off)",
	})

	id, _ := r.Find([]byte("/"))
	if id != 1 {
		t.Errorf("expect matched 1, but %d", id)
	}

	id, _ = r.Find([]byte("/check"))
	if id != 2 {
		t.Errorf("expect matched 2, but %d", id)
	}

	id, _ = r.Find([]byte("/check2"))
	if id != 3 {
		t.Errorf("expect matched 3, but %d", id)
	}

	id, _ = r.Find([]byte("/123"))
	if id != 4 {
		t.Errorf("expect matched 4, but %d", id)
	}

	id, _ = r.Find([]byte("/on"))
	if id != 5 {
		t.Errorf("expect matched 5, but %d", id)
	}
	id, _ = r.Find([]byte("/off"))
	if id != 5 {
		t.Errorf("expect matched 5, but %d", id)
	}

	id, _ = r.Find([]byte("/on/notmatches"))
	if id != 0 {
		t.Errorf("expect not matched , but %d", id)
	}
}
