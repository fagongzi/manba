package forge

import (
	"errors"
	"fmt"
)

// List struct used for holding data neede for Reference data type
type List struct {
	values []Value
}

// NewList will create and initialize a new List value
func NewList() *List {
	return &List{
		values: make([]Value, 0),
	}
}

// GetType will simply return back LIST
func (list *List) GetType() ValueType {
	return LIST
}

// GetValue will resolve and return the value from the underlying list
// this is necessary to inherit from Value
func (list *List) GetValue() interface{} {
	var values []interface{}
	for _, val := range list.values {
		values = append(values, val.GetValue())
	}
	return values
}

// GetValues will return back the list of underlygin values
func (list *List) GetValues() []Value {
	return list.values
}

// UpdateValue will set the underlying list value
func (list *List) UpdateValue(value interface{}) error {
	// Valid types
	switch value.(type) {
	case []Value:
		list.values = value.([]Value)
	default:
		msg := fmt.Sprintf("Unsupported type, %s must be of type []Value", value)
		return errors.New(msg)
	}
	return nil
}

// Get will return the Value at the index
func (list *List) Get(idx int) (Value, error) {
	if idx > list.Length() {
		return nil, errors.New("index out of range")
	}
	return list.values[idx], nil
}

// GetBoolean will try to get the value stored at the index as a bool
// will respond with an error if the value does not exist or cannot be converted to a bool
func (list *List) GetBoolean(idx int) (bool, error) {
	value, err := list.Get(idx)
	if err != nil {
		return false, err
	}

	switch value.(type) {
	case *Primative:
		return value.(*Primative).AsBoolean()
	}

	return false, errors.New("could not convert unknown value to boolean")
}

// GetFloat will try to get the value stored at the index as a float64
// will respond with an error if the value does not exist or cannot be converted to a float64
func (list *List) GetFloat(idx int) (float64, error) {
	value, err := list.Get(idx)
	if err != nil {
		return float64(0), err
	}

	switch value.(type) {
	case *Primative:
		return value.(*Primative).AsFloat()
	}

	return float64(0), errors.New("could not convert non-primative value to float")
}

// GetInteger will try to get the value stored at the index as a int64
// will respond with an error if the value does not exist or cannot be converted to a int64
func (list *List) GetInteger(idx int) (int64, error) {
	value, err := list.Get(idx)
	if err != nil {
		return int64(0), err
	}

	switch value.(type) {
	case *Primative:
		return value.(*Primative).AsInteger()
	}

	return int64(0), errors.New("could not convert non-primative value to integer")
}

// GetList will try to get the value stored at the index as a List
// will respond with an error if the value does not exist or is not a List
func (list *List) GetList(idx int) (*List, error) {
	value, err := list.Get(idx)
	if err != nil {
		return nil, err
	}

	if value.GetType() == LIST {
		return value.(*List), nil
	}

	return nil, errors.New("could not fetch value as list")
}

// GetString will try to get the value stored at the index as a string
// will respond with an error if the value does not exist or cannot be converted to a string
func (list *List) GetString(idx int) (string, error) {
	value, err := list.Get(idx)
	if err != nil {
		return "", err
	}

	switch value.(type) {
	case *Primative:
		return value.(*Primative).AsString()
	}

	return "", errors.New("could not convert non-primative value to string")
}

// Set will set the new Value at the index
func (list *List) Set(idx int, value Value) {
	list.values[idx] = value
}

// Append will append a new Value on the end of the internal list
func (list *List) Append(value Value) {
	list.values = append(list.values, value)
}

// Length will return back the total number of items in the list
func (list *List) Length() int {
	return len(list.values)
}
