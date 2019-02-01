package route

import (
	"testing"
)

func TestNext(t *testing.T) {
	input := []byte("0")
	s := newScanner(input)

	ch := s.Current()
	if ch != '0' {
		t.Errorf("ch expect 0 but %c", ch)
	}

	s.Next()
	ch = s.Current()
	if ch != eoi {
		t.Errorf("ch expect eoi but %c", ch)
	}
}

func TestNextTokenAndScanString(t *testing.T) {
	input := []byte("0/1(2|3:4)")
	s := newScanner(input)

	s.NextToken()
	token := s.Token()
	if token != tokenSlash {
		t.Errorf("token expect / but %d", token)
	}
	value := string(s.ScanString())
	if value != "0" {
		t.Errorf("scanstring expect 0 but %s", value)
	}

	s.NextToken()
	token = s.Token()
	if token != tokenLParen {
		t.Errorf("token expect ( but %d", token)
	}
	value = string(s.ScanString())
	if value != "1" {
		t.Errorf("scanstring expect 1 but %s", value)
	}

	s.NextToken()
	token = s.Token()
	if token != tokenVertical {
		t.Errorf("token expect | but %d", token)
	}
	value = string(s.ScanString())
	if value != "2" {
		t.Errorf("scanstring expect 2 but %s", value)
	}

	s.NextToken()
	token = s.Token()
	if token != tokenColon {
		t.Errorf("token expect : but %d", token)
	}
	value = string(s.ScanString())
	if value != "3" {
		t.Errorf("scanstring expect 3 but %s", value)
	}

	s.NextToken()
	token = s.Token()
	if token != tokenRParen {
		t.Errorf("token expect ) but %d", token)
	}
	value = string(s.ScanString())
	if value != "4" {
		t.Errorf("scanstring expect 4 but %s", value)
	}

	s.NextToken()
	token = s.Token()
	if token != tokenEOF {
		t.Errorf("token expect eof but %d", token)
	}
	value = string(s.ScanString())
	if value != "" {
		t.Errorf("scanstring expect empty but %s", value)
	}
}
