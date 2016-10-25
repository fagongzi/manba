package token

import "fmt"

type Token struct {
	ID      TokenID
	Literal string
	Line    int
	Column  int
}

func (this Token) String() string {
	return fmt.Sprintf(
		"ID<%s> Literal<%s> Line<%s> Column<%s>",
		this.ID, this.Literal, this.Line, this.Column,
	)
}

type TokenID int

const (
	ILLEGAL TokenID = iota
	EOF

	LBRACE
	RBRACE
	LBRACKET
	RBRACKET
	EQUAL
	SEMICOLON
	NEWLINE
	COMMA
	PERIOD

	IDENTIFIER
	BOOLEAN
	INTEGER
	FLOAT
	STRING
	NULL
	COMMENT
	INCLUDE
)

var tokenNames = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	LBRACE:     "LBRACE",
	RBRACE:     "RBRACE",
	LBRACKET:   "LBRACKET",
	RBRACKET:   "RBRACKET",
	EQUAL:      "EQUAL",
	SEMICOLON:  "SEMICOLON",
	NEWLINE:    "NEWLINE",
	COMMA:      "COMMA",
	PERIOD:     "PERIOD",
	IDENTIFIER: "IDENTIFIER",
	BOOLEAN:    "BOOLEAN",
	INTEGER:    "INTEGER",
	FLOAT:      "FLOAT",
	STRING:     "STRING",
	NULL:       "NULL",
	COMMENT:    "COMMENT",
	INCLUDE:    "INCLUDE",
}

func (this TokenID) String() string {
	s := ""
	if 0 <= this && this < TokenID(len(tokenNames)) {
		s = tokenNames[this]
	}

	if s == "" {
		s = "UNKNOWN"
	}

	return s
}
