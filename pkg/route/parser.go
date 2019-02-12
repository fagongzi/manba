package route

import (
	"bytes"
	"fmt"
)

type node struct {
	nt      nodeType
	value   []byte
	enums   [][]byte
	argName []byte
}

func (n *node) isEnum() bool {
	return n.nt == enumType
}

func (n *node) isConst() bool {
	return n.nt == constType
}

func (n *node) addEnum(value []byte) {
	n.enums = append(n.enums, value)
}

func (n *node) setArgName(value []byte) {
	n.argName = value
}

func (n *node) isNumberValue() bool {
	for _, v := range n.value {
		if v < '0' || v > '9' {
			return false
		}
	}

	return true
}

func (n *node) inEnumValue(value []byte) bool {
	in := false
	for _, enum := range n.enums {
		if bytes.Compare(enum, value) == 0 {
			in = true
		}
	}

	return in
}

type parser struct {
	nodes []node
	input []byte
	lexer lexer
}

func newParser(input []byte) *parser {
	if len(input) > 1 && input[len(input)-1] == slash {
		input = input[0 : len(input)-1]
	}

	return &parser{
		input: input,
		lexer: newScanner(input),
	}
}

func (p *parser) parse() ([]node, error) {
	if len(p.input) == 1 && p.input[0] == slash {
		return []node{node{
			nt:    slashType,
			value: slashValue,
		}}, nil
	}

	// /(number|string:const|enum:m1|m2|m3)[:argname]
	prev := tokenUnknown
	prevIndex := -1

	for {
		p.lexer.NextToken()

		token := p.lexer.Token()
		switch token {
		case tokenSlash:
			if prev == tokenSlash {
				if p.lexer.TokenIndex() == prevIndex+1 {
					return nil, fmt.Errorf("syntax error: // not allowed")
				}

				p.nodes = append(p.nodes, node{
					value: p.lexer.ScanString(),
					nt:    constType,
				})
			} else if prev == tokenColon {
				if prevNode := p.prevNode(); prevNode != nil && !prevNode.isConst() {
					prevNode.setArgName(p.lexer.ScanString())
				} else {
					return nil, fmt.Errorf("syntax error: named arg not support const pattern")
				}
			} else if prev == tokenUnknown {
				p.lexer.ScanString()
			}

			p.nodes = append(p.nodes, node{
				nt:    slashType,
				value: slashValue,
			})
			break
		case tokenLParen:
			if prev != tokenSlash {
				return nil, fmt.Errorf("syntax error: ( must after /")
			}

			p.lexer.ScanString()
			break
		case tokenColon:
			if prev == tokenLParen {
				value := p.lexer.ScanString()
				if !bytes.Equal(value, enumValue) {
					return nil, fmt.Errorf("syntax error: not a enum type")
				}

				p.nodes = append(p.nodes, node{
					nt: enumType,
				})
			} else if prev != tokenRParen {
				return nil, fmt.Errorf("syntax error: unexpect : token")
			} else {
				p.lexer.ScanString()
			}

			break
		case tokenVertical:
			prevNode := p.prevNode()

			if (prev == tokenColon || prev == tokenVertical) &&
				prevNode != nil &&
				prevNode.isEnum() {
				prevNode.addEnum(p.lexer.ScanString())
			} else {
				return nil, fmt.Errorf("syntax error: missing : with enum type")
			}

			break
		case tokenRParen:
			if prev == tokenLParen {
				var nt nodeType
				value := p.lexer.ScanString()

				if bytes.Equal(value, stringValue) {
					nt = stringType
				} else if bytes.Equal(value, numberValue) {
					nt = numberType
				} else {
					return nil, fmt.Errorf("syntax error: unknown type: %s", value)
				}

				p.nodes = append(p.nodes, node{
					nt: nt,
				})
			} else if prev == tokenVertical {
				prevNode := p.prevNode()
				if prevNode == nil || !prevNode.isEnum() {
					return nil, fmt.Errorf("syntax error: missing enum")
				}

				prevNode.addEnum(p.lexer.ScanString())
			} else {
				return nil, fmt.Errorf("syntax error: missing (")
			}

			break
		case tokenEOF:
			if prev == tokenSlash {
				p.nodes = append(p.nodes, node{
					value: p.lexer.ScanString(),
					nt:    constType,
				})
			} else if prev == tokenColon {
				if prevNode := p.prevNode(); prevNode != nil && !prevNode.isConst() {
					prevNode.setArgName(p.lexer.ScanString())
				} else {
					return nil, fmt.Errorf("syntax error: named arg not support const pattern")
				}
			} else if prev == tokenUnknown {
				return nil, fmt.Errorf("syntax error: must start with /")
			}

			return p.nodes, nil
		}

		prev = token
		prevIndex = p.lexer.TokenIndex()
	}
}

func (p *parser) prevNode() *node {
	n := len(p.nodes)
	if n == 0 {
		return nil
	}

	return &p.nodes[n-1]
}
