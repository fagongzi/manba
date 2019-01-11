package route

type scanner struct {
	len   int
	input []byte

	token int
	bp    int
	sp    int
	ch    byte
}

func newScanner(input []byte) lexer {
	scan := &scanner{
		len:   len(input),
		input: input,
		bp:    -1,
		sp:    0,
	}

	scan.Next()
	return scan
}

func (scan *scanner) Next() byte {
	scan.bp++

	if scan.bp < scan.len {
		scan.ch = scan.input[scan.bp]
	} else {
		scan.ch = eoi
	}

	return scan.ch
}

func (scan *scanner) NextToken() {
	for {
		switch scan.ch {
		case '/':
			scan.token = tokenSlash
			scan.Next()
			return
		case '(':
			scan.token = tokenLParen
			scan.Next()
			return
		case '|':
			scan.token = tokenVertical
			scan.Next()
			return
		case ':':
			scan.token = tokenColon
			scan.Next()
			return
		case ')':
			scan.token = tokenRParen
			scan.Next()
			return
		case eoi:
			scan.token = tokenEOF
			scan.Next()
			return
		}

		scan.Next()
	}
}

func (scan *scanner) Current() byte {
	return scan.ch
}

func (scan *scanner) Token() int {
	return scan.token
}

func (scan *scanner) TokenIndex() int {
	return scan.bp - 1
}

func (scan *scanner) ScanString() []byte {
	value := scan.input[scan.sp : scan.bp-1]
	scan.sp = scan.bp
	return value
}
