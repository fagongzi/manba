package route

type lexer interface {
	Next() byte
	Current() byte
	NextToken()
	Token() int
	TokenIndex() int
	ScanString() []byte
}
