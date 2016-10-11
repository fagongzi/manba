// Package forge provides an api to deal with parsing configuration files.
//
// Config file example:
//
//     # example.cfg
//     top_level = "a string";
//     primary {
//       primary_int = 500;
//       sub_section {
//         sub_float = 50.5;  # End of line comment
//       }
//     }
//     secondary {
//       secondary_bool = true;
//       secondary_null = null;
//
//       list = [
//         "any",
//         'value',
//         true,
//         55.5,
//       ]
//
//       # Reference other config value
//       local_ref = .secondary_null;
//       global_ref = primary.sub_section.sub_float;
//
//       # Include all files matching the provided pattern
//       include "/etc/app/*.cfg";
//     }
//
//
// Config file format:
//
//     IDENTIFIER: [_a-zA-Z]([_a-zA-Z0-9]+)?
//     NUMBERS: [0-9]+
//     END: ';' | '\n'
//
//     BOOL: 'true' | 'false'
//     NULL: 'null'
//     INTEGER: ('-')? NUMBERS
//     FLOAT: ('-')? NUMBERS '.' NUMBERS
//     STRING: ['"] .* ['"]
//     REFERENCE: (IDENTIFIER)? ('.' IDENTIFIER)+
//     VALUE: BOOL | NULL | INTEGER | FLOAT | STRING | REFERENCE
//     LIST: '[' (VALUE | LIST) (',' NEWLINE* (VALUE | LIST))+ ']'
//
//     INCLUDE: 'include ' STRING END
//     DIRECTIVE: (IDENTIFIER '=' (VALUE | LIST) | INCLUDE) END
//     SECTION: IDENTIFIER '{' (DIRECTIVE | SECTION)* '}'
//     COMMENT: '#' .* '\n'
//
//     CONFIG_FILE: (COMMENT | DIRECTIVE | SECTION)*
//
//
// Values
//  * String:
//      Any value enclosed in double or single quotes (e.g. "string" or 'string').
//      Double quotes, single quotes, and backslashes can be escaped with backslashes (e.g. "\"quoted\"", '\'quoted\'', and "\\<--backslash")
//  * Integer:
//      Any number without decimal places (e.g. 500)
//  * Float:
//      Any number with decimal places (e.g. 500.55)
//  * Boolean:
//      The identifiers 'true' or 'false' of any case (e.g. TRUE, True, true, FALSE, False, false)
//  * Null:
//      The identifier 'null' of any case (e.g. NULL, Null, null)
//  * List:
//      A list value is any number of other values separated by commas and surrounded by brackets.
//      (e.g. [50.5, 'some', "string", true, false])
//  * Global reference:
//      An identifier which may contain periods, the references are resolved from the global
//      section (e.g. global_value, section.sub_section.value)
//  * Local reference:
//      An identifier which main contain periods which starts with a period, the references
//      are resolved from the settings current section (e.g. .value, .sub_section.value)
//
// Directives
//  * Comment:
//      A comment is a pound symbol ('#') followed by any text any which ends with a newline (e.g. '# I am a comment\n')
//      A comment can either be on a line of it's own or at the end of any line. Nothing can come after the comment
//      until after the newline.
//  * Directive:
//      A directive is a setting, a identifier and a value. They are in the format '<identifier> = <value>;'
//      All directives must end in either a semicolon or newline. The value can be any of the types defined above.
//  * Section:
//      A section is a grouping of directives under a common name. They are in the format '<section_name> { <directives> }'.
//      All sections must be wrapped in braces ('{', '}') and must all have a name. They do not end in a semicolon.
//      Sections may be left empty, they do not have to contain any directives.
//  * Include:
//      An include statement tells the config parser to include the contents of another config file where the include
//      statement is defined. Includes are in the format 'include "<pattern>";'. The <pattern> can be any glob
//      like pattern which is compatible with `path.filepath.Match` http://golang.org/pkg/path/filepath/#Match
//
package forge

import (
	"bytes"
	"io"
	"strings"
)

// ParseBytes takes a []byte representation of the config file, parses it
// and responds with `*Section` and potentially an `error` if it cannot
// properly parse the config
func ParseBytes(data []byte) (*Section, error) {
	return ParseReader(bytes.NewReader(data))
}

// ParseFile takes a string filename for the config file, parses it
// and responds with `*Section` and potentially an `error` if it cannot
// properly parse the configf
func ParseFile(filename string) (*Section, error) {
	parser, err := NewFileParser(filename)
	if err != nil {
		return nil, err
	}
	err = parser.Parse()
	if err != nil {
		return nil, err
	}

	return parser.GetSettings(), nil
}

// ParseReader takes an `io.Reader` representation of the config file, parses it
// and responds with `*Section` and potentially an `error` if it cannot
// properly parse the config
func ParseReader(reader io.Reader) (*Section, error) {
	parser := NewParser(reader)
	err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return parser.GetSettings(), nil
}

// ParseString takes a string representation of the config file, parses it
// and responds with `*Section` and potentially an `error` if it cannot
// properly parse the config
func ParseString(data string) (*Section, error) {
	return ParseReader(strings.NewReader(data))
}
