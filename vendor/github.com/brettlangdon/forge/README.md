forge
=====

[![Build Status](https://travis-ci.org/brettlangdon/forge.svg?branch=master)](https://travis-ci.org/brettlangdon/forge)
[![GoDoc](https://godoc.org/github.com/brettlangdon/forge?status.svg)](https://godoc.org/github.com/brettlangdon/forge)

Forge is a configuration syntax and parser.

## Installation

`go get github.com/brettlangdon/forge`

## Documentation

Documentation can be viewed on godoc: https://godoc.org/github.com/brettlangdon/forge

## Example

You can see example usage in the `example` folder.

```cfg
# example.cfg

# Global directives
global = "global value";
# Primary section
primary {
  string = "primary string value";
  single = 'single quotes are allowed too';

  # Semicolons are optional
  integer = 500
  float = 80.80
  boolean = true
  negative = FALSE
  nothing = NULL

  list = [50.5, true, false, "hello", 'world'];

  # Include external files
  include "./include*.cfg";
  # Primary-sub section
  sub {
      key = "primary sub key value";
  }
}

# Secondary section
secondary {
  another = "secondary another value";
  global_reference = global;
  primary_sub_key = primary.sub.key;
  another_again = .another;  # References secondary.another
  _under = 50;
}
```

```go
package main

import (
	"fmt"
	"json"

	"github.com/brettlangdon/forge"
)

func main() {
	// Parse a `SectionValue` from `example.cfg`
	settings, err := forge.ParseFile("example.cfg")
	if err != nil {
		panic(err)
	}

	// Get a single value
	if settings.Exists("global") {
		// Get `global` casted as a string
		value, _ := settings.GetString("global")
		fmt.Printf("global = \"%s\"\r\n", value)
	}

	// Get a nested value
	value, err := settings.Resolve("primary.included_setting")
	fmt.Printf("primary.included_setting = \"%s\"\r\n", value.GetValue())

	// You can also traverse down the sections manually
	primary, err := settings.GetSection("primary")
	strVal, err := primary.GetString("included_setting")
	fmt.Printf("primary.included_setting = \"%s\"\r\n", strVal)

	// Convert settings to a map
	settingsMap := settings.ToMap()
	fmt.Printf("global = \"%s\"\r\n", settingsMap["global"])

	// Convert settings to JSON
	jsonBytes, err := settings.ToJSON()
	fmt.Printf("\r\n\r\n%s\r\n", string(jsonBytes))
}
```

## Issues/Requests?

Please feel free to open a github issue for any issues you have or any feature requests.
