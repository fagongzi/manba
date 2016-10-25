package forge

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/brettlangdon/forge/token"
)

func isSemicolonOrNewline(id token.TokenID) bool {
	return id == token.SEMICOLON || id == token.NEWLINE
}

// Parser is a struct to hold data necessary for parsing a config from a scanner
type Parser struct {
	files      []string
	settings   *Section
	scanner    *Scanner
	curTok     token.Token
	curSection *Section
	previous   []*Section
}

// NewParser will create and initialize a new Parser from a provided io.Reader
func NewParser(reader io.Reader) *Parser {
	settings := NewSection()
	return &Parser{
		files:      make([]string, 0),
		scanner:    NewScanner(reader),
		settings:   settings,
		curSection: settings,
		previous:   make([]*Section, 0),
	}
}

// NewFileParser will create and initialize a new Parser from a provided from a filename string
func NewFileParser(filename string) (*Parser, error) {
	reader, err := os.Open(filename)
	defer reader.Close()
	if err != nil {
		return nil, err
	}
	parser := NewParser(reader)
	parser.addFile(filename)
	return parser, nil
}

func (parser *Parser) addFile(filename string) {
	parser.files = append(parser.files, filename)
}

func (parser *Parser) hasParsed(search string) bool {
	for _, filename := range parser.files {
		if filename == search {
			return true
		}
	}
	return false
}

func (parser *Parser) syntaxError(msg string) error {
	msg = fmt.Sprintf(
		"syntax error line <%d> column <%d>: %s",
		parser.curTok.Line,
		parser.curTok.Column,
		msg,
	)
	return errors.New(msg)
}

func (parser *Parser) readToken() token.Token {
	parser.curTok = parser.scanner.NextToken()
	return parser.curTok
}

func (parser *Parser) skipNewlines() {
	for parser.curTok.ID == token.NEWLINE {
		parser.readToken()
	}
}

func (parser *Parser) parseList() ([]Value, error) {
	var values []Value
	for {
		parser.skipNewlines()

		value, err := parser.parseSettingValue()
		if err != nil {
			return nil, err
		}
		values = append(values, value)

		if parser.curTok.ID == token.COMMA {
			parser.readToken()
		}

		parser.skipNewlines()

		if parser.curTok.ID == token.RBRACKET {
			parser.readToken()
			break
		}
	}

	return values, nil
}

func (parser *Parser) parseReference(startingSection *Section, period bool) (Value, error) {
	name := ""
	if period == false {
		name = parser.curTok.Literal
	}
	for {
		parser.readToken()
		if parser.curTok.ID == token.PERIOD && period == false {
			period = true
		} else if period && parser.curTok.ID == token.IDENTIFIER {
			if len(name) > 0 {
				name += "."
			}
			name += parser.curTok.Literal
			period = false
		} else if isSemicolonOrNewline(parser.curTok.ID) {
			break
		} else {
			msg := fmt.Sprintf("expected ';' or '\n' instead found '%s'", parser.curTok.Literal)
			return nil, parser.syntaxError(msg)
		}
	}
	if len(name) == 0 {
		return nil, parser.syntaxError(
			fmt.Sprintf("expected IDENTIFIER instead found %s", parser.curTok.Literal),
		)
	}

	if period {
		return nil, parser.syntaxError(fmt.Sprintf("expected IDENTIFIER after PERIOD"))
	}

	return NewReference(name, startingSection), nil
}

func (parser *Parser) parseSettingValue() (Value, error) {
	var value Value

	readNext := true
	switch parser.curTok.ID {
	case token.STRING:
		value = NewString(parser.curTok.Literal)
	case token.BOOLEAN:
		boolVal, err := strconv.ParseBool(parser.curTok.Literal)
		if err != nil {
			return value, nil
		}
		value = NewBoolean(boolVal)
	case token.NULL:
		value = NewNull()
	case token.INTEGER:
		intVal, err := strconv.ParseInt(parser.curTok.Literal, 10, 64)
		if err != nil {
			return value, err
		}
		value = NewInteger(intVal)
	case token.FLOAT:
		floatVal, err := strconv.ParseFloat(parser.curTok.Literal, 64)
		if err != nil {
			return value, err
		}
		value = NewFloat(floatVal)
	case token.PERIOD:
		reference, err := parser.parseReference(parser.curSection, true)
		if err != nil {
			return value, err
		}
		value = reference
		readNext = false
	case token.IDENTIFIER:
		reference, err := parser.parseReference(parser.settings, false)
		if err != nil {
			return value, err
		}
		value = reference
		readNext = false
	case token.LBRACKET:
		parser.readToken()
		listVal, err := parser.parseList()
		if err != nil {
			return value, err
		}
		value = NewList()
		value.UpdateValue(listVal)
		readNext = false
	default:
		return value, parser.syntaxError(
			fmt.Sprintf("expected STRING, INTEGER, FLOAT, BOOLEAN or IDENTIFIER, instead found %s", parser.curTok.ID),
		)
	}

	if readNext {
		parser.readToken()
	}
	return value, nil
}

func (parser *Parser) parseSetting(name string) error {
	parser.readToken()
	value, err := parser.parseSettingValue()
	if err != nil {
		return err
	}
	if isSemicolonOrNewline(parser.curTok.ID) == false {
		msg := fmt.Sprintf("expected ';' or '\n' instead found '%s'", parser.curTok.Literal)
		return parser.syntaxError(msg)
	}
	parser.readToken()

	parser.curSection.Set(name, value)
	return nil
}

func (parser *Parser) parseInclude() error {
	if parser.curTok.ID != token.STRING {
		msg := fmt.Sprintf("expected STRING instead found '%s'", parser.curTok.ID)
		return parser.syntaxError(msg)
	}
	pattern := parser.curTok.Literal

	parser.readToken()
	if isSemicolonOrNewline(parser.curTok.ID) == false {
		msg := fmt.Sprintf("expected ';' or '\n' instead found '%s'", parser.curTok.Literal)
		return parser.syntaxError(msg)
	}

	filenames, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	oldScanner := parser.scanner
	for _, filename := range filenames {
		// We have already visited this file, don't include again
		// DEV: This can cause recursive includes if this isn't here :o
		if parser.hasParsed(filename) {
			continue
		}
		reader, err := os.Open(filename)
		defer reader.Close()
		if err != nil {
			return err
		}
		parser.curSection.AddInclude(filename)
		parser.scanner = NewScanner(reader)
		parser.parse()
		// Make sure to add the filename to the internal list to ensure we don't
		// accidentally recursively include config files
		parser.addFile(filename)
	}
	parser.scanner = oldScanner
	parser.readToken()
	return nil
}

func (parser *Parser) parseSection(name string) error {
	section := parser.curSection.AddSection(name)
	parser.previous = append(parser.previous, parser.curSection)
	parser.curSection = section
	return nil
}

func (parser *Parser) endSection() error {
	if len(parser.previous) == 0 {
		return parser.syntaxError("unexpected section end '}'")
	}

	pLen := len(parser.previous)
	previous := parser.previous[pLen-1]
	parser.previous = parser.previous[0 : pLen-1]
	parser.curSection = previous
	return nil
}

func (parser *Parser) parse() error {
	parser.readToken()
	for {
		if parser.curTok.ID == token.EOF {
			break
		}
		tok := parser.curTok
		parser.readToken()
		switch tok.ID {
		case token.COMMENT:
			parser.curSection.AddComment(tok.Literal)
		case token.INCLUDE:
			parser.parseInclude()
		case token.IDENTIFIER:
			if parser.curTok.ID == token.LBRACE {
				err := parser.parseSection(tok.Literal)
				if err != nil {
					return err
				}
				parser.readToken()
			} else if parser.curTok.ID == token.EQUAL {
				err := parser.parseSetting(tok.Literal)
				if err != nil {
					return err
				}
			}
		case token.RBRACE:
			err := parser.endSection()
			if err != nil {
				return err
			}
		case token.NEWLINE:
			// Ignore extra newlines
			continue
		default:
			return parser.syntaxError(fmt.Sprintf("unexpected token %s", tok))
		}
	}
	return nil
}

// GetSettings will fetch the parsed settings from this Parser
func (parser *Parser) GetSettings() *Section {
	return parser.settings
}

// Parse will tell the Parser to parse all settings from the config
func (parser *Parser) Parse() error {
	err := parser.parse()
	if err != nil {
		return err
	}

	if len(parser.previous) > 0 {
		return parser.syntaxError("expected end of section, instead found EOF")
	}

	return nil
}
