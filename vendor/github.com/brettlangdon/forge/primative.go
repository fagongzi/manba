package forge

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// Primative struct for holding data about primative values
type Primative struct {
	valueType ValueType
	value     interface{}
}

func newPrimative(valueType ValueType, value interface{}) *Primative {
	return &Primative{
		valueType: valueType,
		value:     value,
	}
}

// NewPrimative will create a new empty Primative and call UpdateValue
// with the provided value
func NewPrimative(value interface{}) (*Primative, error) {
	primative := &Primative{}
	err := primative.UpdateValue(value)
	return primative, err
}

// NewBoolean will create and initialize a new Boolean type primative value
func NewBoolean(value bool) *Primative {
	return newPrimative(BOOLEAN, value)
}

// NewFloat will create and initialize a new Float type primative value
func NewFloat(value float64) *Primative {
	return newPrimative(FLOAT, value)
}

// NewInteger will create and initialize a new Integer type primative value
func NewInteger(value int64) *Primative {
	return newPrimative(INTEGER, value)
}

// NewNull will create and initialize a new Null type primative value
func NewNull() *Primative {
	return newPrimative(NULL, nil)
}

// NewString will create and initialize a new String type primative value
func NewString(value string) *Primative {
	return newPrimative(STRING, value)
}

// GetType will return the ValueType associated with this primative
func (primative *Primative) GetType() ValueType {
	return primative.valueType
}

// GetValue returns the raw interface{} value assigned to this primative
func (primative *Primative) GetValue() interface{} {
	return primative.value
}

// UpdateValue will update the internal value and stored ValueType for this primative
func (primative *Primative) UpdateValue(value interface{}) error {
	// Valid types
	switch value.(type) {
	case bool:
		primative.valueType = BOOLEAN
	case float64:
		primative.valueType = FLOAT
	case int:
		value = int64(value.(int))
		primative.valueType = INTEGER
	case int64:
		primative.valueType = INTEGER
	case nil:
		primative.valueType = NULL
	case string:
		primative.valueType = STRING
	default:
		msg := fmt.Sprintf("Unsupported type, %s must be of (bool, float64, int64, nil, string)", value)
		return errors.New(msg)

	}
	primative.value = value
	return nil
}

// AsBoolean tries to convert/return the value stored in this primative as a bool
func (primative *Primative) AsBoolean() (bool, error) {
	switch val := primative.value.(type) {
	case bool:
		return val, nil
	case float64:
		return val != 0, nil
	case int64:
		return val != 0, nil
	case nil:
		return false, nil
	case string:
		return val != "", nil
	}

	msg := fmt.Sprintf("Could not convert value %s to type BOOLEAN", primative.value)
	return false, errors.New(msg)
}

// AsFloat tries to convert/return the value stored in this primative as a float64
func (primative *Primative) AsFloat() (float64, error) {
	switch val := primative.value.(type) {
	case bool:
		floatVal := float64(0)
		if val {
			floatVal = float64(1)
		}

		return floatVal, nil
	case float64:
		return val, nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	}

	msg := fmt.Sprintf("Could not convert value %s to type FLOAT", primative.value)
	return 0, errors.New(msg)
}

// AsInteger tries to convert/return the value stored in this primative as a int64
func (primative *Primative) AsInteger() (int64, error) {
	switch val := primative.value.(type) {
	case bool:
		intVal := int64(0)
		if val {
			intVal = int64(1)
		}
		return intVal, nil
	case float64:
		return int64(math.Trunc(val)), nil
	case int64:
		return val, nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	}

	msg := fmt.Sprintf("Could not convert value %s to type INTEGER", primative.value)
	return 0, errors.New(msg)
}

// AsNull tries to convert/return the value stored in this primative as a null
func (primative *Primative) AsNull() (interface{}, error) {
	switch val := primative.value.(type) {
	case nil:
		return val, nil
	}

	msg := fmt.Sprintf("Could not convert value %s to nil", primative.value)
	return 0, errors.New(msg)
}

// AsString tries to convert/return the value stored in this primative as a string
func (primative *Primative) AsString() (string, error) {
	switch val := primative.value.(type) {
	case bool:
		strVal := "False"
		if val {
			strVal = "True"
		}
		return strVal, nil
	case float64:
		return strconv.FormatFloat(val, 10, -1, 64), nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	case nil:
		return "Null", nil
	case string:
		return val, nil
	}

	msg := fmt.Sprintf("Could not convert value %s to type STRING", primative.value)
	return "", errors.New(msg)
}

func (primative *Primative) String() string {
	str, _ := primative.AsString()
	return str
}
