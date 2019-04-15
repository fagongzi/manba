package route

const (
	eoi      byte = 0x1A
	slash         = byte('/')
	lParen        = byte('(')
	rParen        = byte(')')
	vertical      = byte('|')
	colon         = byte(':')
)

const (
	tokenEOF = iota
	tokenUnknown
	tokenSlash
	tokenLParen
	tokenRParen
	tokenVertical
	tokenColon
)

var (
	slashValue  = []byte("/")
	numberValue = []byte("number")
	stringValue = []byte("string")
	enumValue   = []byte("enum")

	matchAll = []byte("*")
)

type nodeType int

const (
	slashType  = nodeType(5)
	stringType = nodeType(4)
	constType  = nodeType(3)
	enumType   = nodeType(2)
	numberType = nodeType(1)
)
