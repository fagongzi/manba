package forge

// ValueType is an int type for representing the types of values forge can handle
type ValueType int

const (
	// UNKNOWN ValueType
	UNKNOWN ValueType = iota

	primativesStart
	// BOOLEAN ValueType
	BOOLEAN
	// FLOAT ValueType
	FLOAT
	// INTEGER ValueType
	INTEGER
	// NULL ValueType
	NULL
	// STRING ValueType
	STRING
	primativesDnd

	complexStart
	// LIST ValueType
	LIST
	// REFERENCE ValueType
	REFERENCE
	// SECTION ValueType
	SECTION
	complexEnd
)

var valueTypes = [...]string{
	BOOLEAN: "BOOLEAN",
	FLOAT:   "FLOAT",
	INTEGER: "INTEGER",
	NULL:    "NULL",
	STRING:  "STRING",

	LIST:      "LIST",
	REFERENCE: "REFERENCE",
	SECTION:   "SECTION",
}

func (valueType ValueType) String() string {
	str := ""
	if 0 <= valueType && valueType < ValueType(len(valueTypes)) {
		str = valueTypes[valueType]
	}

	if str == "" {
		str = "UNKNOWN"
	}

	return str
}

// Value is the base interface for Primative and Section data types
type Value interface {
	GetType() ValueType
	GetValue() interface{}
	UpdateValue(interface{}) error
}
