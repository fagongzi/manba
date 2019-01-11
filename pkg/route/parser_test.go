package route

import (
	"bytes"
	"testing"
)

func TestParseSingle(t *testing.T) {
	input := []byte("/")
	p := newParser(input)
	nodes, err := p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 1 {
		t.Errorf("expect 1 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
}

func TestParseConst(t *testing.T) {
	input := []byte("/abc")
	p := newParser(input)
	nodes, err := p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 2 {
		t.Errorf("expect 2 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("abc")) {
		t.Errorf("expect value abc at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != constType {
		t.Errorf("expect node type %d at index 1 but %d", constType, nodes[1].nt)
	}

	input = []byte("/abc/")
	p = newParser(input)
	nodes, err = p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 2 {
		t.Errorf("expect 2 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("abc")) {
		t.Errorf("expect value abc at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != constType {
		t.Errorf("expect node type %d at index 1 but %d", constType, nodes[1].nt)
	}

	input = []byte("/abc/def")
	p = newParser(input)
	nodes, err = p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 4 {
		t.Errorf("expect 4 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("abc")) {
		t.Errorf("expect value abc at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != constType {
		t.Errorf("expect node type %d at index 1 but %d", constType, nodes[1].nt)
	}
	if !bytes.Equal(nodes[2].value, []byte("/")) {
		t.Errorf("expect value / at index 2 but %s", nodes[2].value)
	}
	if nodes[2].nt != slashType {
		t.Errorf("expect node type %d at index 2 but %d", slashType, nodes[2].nt)
	}
	if !bytes.Equal(nodes[3].value, []byte("def")) {
		t.Errorf("expect value abc at index 3 but %s", nodes[3].value)
	}
	if nodes[3].nt != constType {
		t.Errorf("expect node type %d at index 3 but %d", constType, nodes[3].nt)
	}
}

func TestParseNumber(t *testing.T) {
	input := []byte("/(number)")
	p := newParser(input)
	nodes, err := p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 2 {
		t.Errorf("expect 2 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("")) {
		t.Errorf("expect value  at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != numberType {
		t.Errorf("expect node type %d at index 1 but %d", numberType, nodes[1].nt)
	}

	input = []byte("/(number):age")
	p = newParser(input)
	nodes, err = p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 2 {
		t.Errorf("expect 2 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("")) {
		t.Errorf("expect value  at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != numberType {
		t.Errorf("expect node type %d at index 1 but %d", numberType, nodes[1].nt)
	}
	if string(nodes[1].argName) != "age" {
		t.Errorf("expect node arg name age at index 1 but %s", nodes[1].argName)
	}
}

func TestParseString(t *testing.T) {
	input := []byte("/(string)")
	p := newParser(input)
	nodes, err := p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 2 {
		t.Errorf("expect 2 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("")) {
		t.Errorf("expect value  at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != stringType {
		t.Errorf("expect node type %d at index 1 but %d", stringType, nodes[1].nt)
	}

	input = []byte("/(string):age")
	p = newParser(input)
	nodes, err = p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 2 {
		t.Errorf("expect 2 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("")) {
		t.Errorf("expect value  at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != stringType {
		t.Errorf("expect node type %d at index 1 but %d", stringType, nodes[1].nt)
	}
	if string(nodes[1].argName) != "age" {
		t.Errorf("expect node arg name age at index 1 but %s", nodes[1].argName)
	}
}

func TestParseEnum(t *testing.T) {
	input := []byte("/(enum:m1|m2|m3)")
	p := newParser(input)
	nodes, err := p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 2 {
		t.Errorf("expect 2 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("")) {
		t.Errorf("expect value  at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != enumType {
		t.Errorf("expect node type %d at index 1 but %d", enumType, nodes[1].nt)
	}
	if len(nodes[1].enums) != 3 {
		t.Errorf("expect enums 3 at index 1 but %d", len(nodes[1].enums))
	}
	if string(nodes[1].enums[0]) != "m1" ||
		string(nodes[1].enums[1]) != "m2" ||
		string(nodes[1].enums[2]) != "m3" {
		t.Errorf("expect enums m1|m2|m3 at index 1 but %s|%s|%s",
			nodes[1].enums[0],
			nodes[1].enums[1],
			nodes[1].enums[2])
	}

	input = []byte("/(enum:m1|m2|m3):age")
	p = newParser(input)
	nodes, err = p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}
	if len(nodes) != 2 {
		t.Errorf("expect 2 nodes but %d", len(nodes))
	}
	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}
	if !bytes.Equal(nodes[1].value, []byte("")) {
		t.Errorf("expect value  at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != enumType {
		t.Errorf("expect node type %d at index 1 but %d", enumType, nodes[1].nt)
	}
	if len(nodes[1].enums) != 3 {
		t.Errorf("expect enums 3 at index 1 but %d", len(nodes[1].enums))
	}
	if string(nodes[1].enums[0]) != "m1" ||
		string(nodes[1].enums[1]) != "m2" ||
		string(nodes[1].enums[2]) != "m3" {
		t.Errorf("expect enums m1|m2|m3 at index 1 but %s|%s|%s",
			nodes[1].enums[0],
			nodes[1].enums[1],
			nodes[1].enums[2])
	}
	if string(nodes[1].argName) != "age" {
		t.Errorf("expect node arg name age at index 1 but %s", nodes[1].argName)
	}
}

func TestParseComplex(t *testing.T) {
	input := []byte("/api/(string):version/(number):id/(enum:on|off):action")
	p := newParser(input)
	nodes, err := p.parse()
	if err != nil {
		t.Errorf("parsed failed with error: %+v", err)
		return
	}

	if len(nodes) != 8 {
		t.Errorf("expect 8 nodes but %d", len(nodes))
	}

	if !bytes.Equal(nodes[0].value, []byte("/")) {
		t.Errorf("expect value / at index 0 but %s", nodes[0].value)
	}
	if nodes[0].nt != slashType {
		t.Errorf("expect node type %d at index 0 but %d", slashType, nodes[0].nt)
	}

	if !bytes.Equal(nodes[1].value, []byte("api")) {
		t.Errorf("expect value api at index 1 but %s", nodes[1].value)
	}
	if nodes[1].nt != constType {
		t.Errorf("expect node type %d at index 1 but %d", constType, nodes[1].nt)
	}

	if !bytes.Equal(nodes[2].value, []byte("/")) {
		t.Errorf("expect value / at index 2 but %s", nodes[2].value)
	}
	if nodes[2].nt != slashType {
		t.Errorf("expect node type %d at index 2 but %d", slashType, nodes[2].nt)
	}

	if !bytes.Equal(nodes[3].value, []byte("")) {
		t.Errorf("expect value empty at index 3 but %s", nodes[3].value)
	}
	if nodes[3].nt != stringType {
		t.Errorf("expect node type %d at index 3 but %d", stringType, nodes[3].nt)
	}
	if string(nodes[3].argName) != "version" {
		t.Errorf("expect node arg name version at index 3 but %s", nodes[3].argName)
	}

	if !bytes.Equal(nodes[4].value, []byte("/")) {
		t.Errorf("expect value / at index 4 but %s", nodes[4].value)
	}
	if nodes[4].nt != slashType {
		t.Errorf("expect node type %d at index 4 but %d", slashType, nodes[4].nt)
	}

	if !bytes.Equal(nodes[5].value, []byte("")) {
		t.Errorf("expect value empty at index 5 but %s", nodes[5].value)
	}
	if nodes[5].nt != numberType {
		t.Errorf("expect node type %d at index 5 but %d", numberType, nodes[5].nt)
	}
	if string(nodes[5].argName) != "id" {
		t.Errorf("expect node arg name id at index 3 but %s", nodes[5].argName)
	}

	if !bytes.Equal(nodes[6].value, []byte("/")) {
		t.Errorf("expect value / at index 6 but %s", nodes[6].value)
	}
	if nodes[6].nt != slashType {
		t.Errorf("expect node type %d at index 6 but %d", slashType, nodes[6].nt)
	}

	if !bytes.Equal(nodes[7].value, []byte("")) {
		t.Errorf("expect value empty at index 7 but %s", nodes[7].value)
	}
	if nodes[7].nt != enumType {
		t.Errorf("expect node type %d at index 7 but %d", enumType, nodes[7].nt)
	}
	if string(nodes[7].argName) != "action" {
		t.Errorf("expect node arg name action at index 7 but %s", nodes[7].argName)
	}
	if len(nodes[7].enums) != 2 {
		t.Errorf("expect enums 2 at index 7 but %d", len(nodes[7].enums))
	}
	if string(nodes[7].enums[0]) != "on" ||
		string(nodes[7].enums[1]) != "off" {
		t.Errorf("expect enums on|off at index 7 but %s|%s",
			nodes[7].enums[0],
			nodes[7].enums[1])
	}
}
