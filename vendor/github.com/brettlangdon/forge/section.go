package forge

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrNotExists represents a nonexistent value error
	ErrNotExists = errors.New("value does not exist")
)

// Section struct holds a map of values
type Section struct {
	comments []string
	includes []string
	parent   *Section
	values   map[string]Value
}

// NewSection will create and initialize a new Section
func NewSection() *Section {
	return &Section{
		comments: make([]string, 0),
		includes: make([]string, 0),
		values:   make(map[string]Value),
	}
}

func newChildSection(parent *Section) *Section {
	return &Section{
		comments: make([]string, 0),
		includes: make([]string, 0),
		parent:   parent,
		values:   make(map[string]Value),
	}
}

// AddComment will append a new comment into the section
func (section *Section) AddComment(comment string) {
	section.comments = append(section.comments, comment)
}

// AddInclude will append a new filename into the section
func (section *Section) AddInclude(filename string) {
	section.includes = append(section.includes, filename)
}

// GetComments will return all the comments were defined for this Section
func (section *Section) GetComments() []string {
	return section.comments
}

// GetIncludes will return the filenames of all the includes were parsed for this Section
func (section *Section) GetIncludes() []string {
	return section.includes
}

// GetType will respond with the ValueType of this Section (hint, always SECTION)
func (section *Section) GetType() ValueType {
	return SECTION
}

// GetValue retrieves the raw underlying value stored in this Section
func (section *Section) GetValue() interface{} {
	return section.values
}

// UpdateValue updates the raw underlying value stored in this Section
func (section *Section) UpdateValue(value interface{}) error {
	switch value.(type) {
	case map[string]Value:
		section.values = value.(map[string]Value)
		return nil
	}

	msg := fmt.Sprintf("unsupported type, %s must be of type `map[string]Value`", value)
	return errors.New(msg)
}

// AddSection adds a new child section to this Section with the provided name
func (section *Section) AddSection(name string) *Section {
	childSection := newChildSection(section)
	section.values[name] = childSection
	return childSection
}

// Exists returns true when a value stored under the key exists
func (section *Section) Exists(name string) bool {
	_, err := section.Get(name)
	return err == nil
}

// Get the value (Primative or Section) stored under the name
// will respond with an error if the value does not exist
func (section *Section) Get(name string) (Value, error) {
	value, ok := section.values[name]
	var err error
	if ok == false {
		err = ErrNotExists
	}
	return value, err
}

// GetBoolean will try to get the value stored under name as a bool
// will respond with an error if the value does not exist or cannot be converted to a bool
func (section *Section) GetBoolean(name string) (bool, error) {
	value, err := section.Get(name)
	if err != nil {
		return false, err
	}

	switch value.(type) {
	case *Primative:
		return value.(*Primative).AsBoolean()
	case *Section:
		return true, nil
	}

	return false, errors.New("could not convert unknown value to boolean")
}

// GetFloat will try to get the value stored under name as a float64
// will respond with an error if the value does not exist or cannot be converted to a float64
func (section *Section) GetFloat(name string) (float64, error) {
	value, err := section.Get(name)
	if err != nil {
		return float64(0), err
	}

	switch value.(type) {
	case *Primative:
		return value.(*Primative).AsFloat()
	}

	return float64(0), errors.New("could not convert non-primative value to float")
}

// GetInteger will try to get the value stored under name as a int64
// will respond with an error if the value does not exist or cannot be converted to a int64
func (section *Section) GetInteger(name string) (int64, error) {
	value, err := section.Get(name)
	if err != nil {
		return int64(0), err
	}

	switch value.(type) {
	case *Primative:
		return value.(*Primative).AsInteger()
	}

	return int64(0), errors.New("could not convert non-primative value to integer")
}

// GetList will try to get the value stored under name as a List
// will respond with an error if the value does not exist or is not a List
func (section *Section) GetList(name string) (*List, error) {
	value, err := section.Get(name)
	if err != nil {
		return nil, err
	}

	if value.GetType() == LIST {
		return value.(*List), nil
	}

	return nil, errors.New("could not fetch value as list")
}

// GetSection will try to get the value stored under name as a Section
// will respond with an error if the value does not exist or is not a Section
func (section *Section) GetSection(name string) (*Section, error) {
	value, err := section.Get(name)
	if err != nil {
		return nil, err
	}

	if value.GetType() == SECTION {
		return value.(*Section), nil
	}
	return nil, errors.New("could not fetch value as section")
}

// GetString will try to get the value stored under name as a string
// will respond with an error if the value does not exist or cannot be converted to a string
func (section *Section) GetString(name string) (string, error) {
	value, err := section.Get(name)
	if err != nil {
		return "", err
	}

	switch value.(type) {
	case *Primative:
		return value.(*Primative).AsString()
	}

	return "", errors.New("could not convert non-primative value to string")
}

// GetParent will get the parent section associated with this Section or nil
// if it does not have one
func (section *Section) GetParent() *Section {
	return section.parent
}

// HasParent will return true if this Section has a parent
func (section *Section) HasParent() bool {
	return section.parent != nil
}

// Keys will return back a list of all setting names in this Section
func (section *Section) Keys() []string {
	var keys []string
	for key := range section.values {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

// Set will set a value (Primative or Section) to the provided name
func (section *Section) Set(name string, value Value) {
	section.values[name] = value
}

// SetBoolean will set the value for name as a bool
func (section *Section) SetBoolean(name string, value bool) {
	current, err := section.Get(name)

	// Exists just update the value/type
	if err == nil {
		current.UpdateValue(value)
	} else {
		section.values[name] = NewBoolean(value)
	}
}

// SetFloat will set the value for name as a float64
func (section *Section) SetFloat(name string, value float64) {
	current, err := section.Get(name)

	// Exists just update the value/type
	if err == nil {
		current.UpdateValue(value)
	} else {
		section.values[name] = NewFloat(value)
	}
}

// SetInteger will set the value for name as a int64
func (section *Section) SetInteger(name string, value int64) {
	current, err := section.Get(name)

	// Exists just update the value/type
	if err == nil {
		current.UpdateValue(value)
	} else {
		section.values[name] = NewInteger(value)
	}
}

// SetNull will set the value for name as nil
func (section *Section) SetNull(name string) {
	current, err := section.Get(name)

	// Already is a Null, nothing to do
	if err == nil && current.GetType() == NULL {
		return
	}
	section.Set(name, NewNull())
}

// SetString will set the value for name as a string
func (section *Section) SetString(name string, value string) {
	current, err := section.Get(name)

	// Exists just update the value/type
	if err == nil {
		current.UpdateValue(value)
	} else {
		section.Set(name, NewString(value))
	}
}

// Resolve will recursively try to fetch the provided value and will respond
// with an error if the name does not exist or tries to be resolved through
// a non-section value
func (section *Section) Resolve(name string) (Value, error) {
	// Used only in error state return value
	var value Value

	parts := strings.Split(name, ".")
	if len(parts) == 0 {
		return value, errors.New("no name provided")
	}

	var current Value
	current = section
	for _, part := range parts {
		if current.GetType() != SECTION {
			return value, errors.New("trying to resolve value from non-section")
		}

		nextCurrent, err := current.(*Section).Get(part)
		if err != nil {
			return value, errors.New("could not find value in section")
		}
		current = nextCurrent
	}
	return current, nil
}

// Merge merges the given section to current section. Settings from source
// section overwites the values in the current section
func (section *Section) Merge(source *Section) error {
	for _, key := range source.Keys() {
		sourceValue, _ := source.Get(key)
		targetValue, err := section.Get(key)

		// not found, so add it
		if err != nil {
			section.Set(key, sourceValue)
			continue
		}

		// found existing one and it's type SECTION, merge it
		if targetValue.GetType() == SECTION {
			// Source value have to be SECTION type here
			if sourceValue.GetType() != SECTION {
				return fmt.Errorf("source (%v) and target (%v) type doesn't match: %v",
					sourceValue.GetType(),
					targetValue.GetType(),
					key)
			}

			if err = targetValue.(*Section).Merge(sourceValue.(*Section)); err != nil {
				return err
			}

			continue
		}

		// found existing one, update it
		if err = targetValue.UpdateValue(sourceValue.GetValue()); err != nil {
			return fmt.Errorf("%v: %v", err, key)
		}
	}
	return nil
}

// ToJSON will convert this Section and all it's underlying values and Sections
// into JSON as a []byte
func (section *Section) ToJSON() ([]byte, error) {
	data := section.ToMap()
	return json.Marshal(data)
}

// ToMap will convert this Section and all it's underlying values and Sections into
// a map[string]interface{}
func (section *Section) ToMap() map[string]interface{} {
	output := make(map[string]interface{})

	for key, value := range section.values {
		if value.GetType() == SECTION {
			output[key] = value.(*Section).ToMap()
		} else {
			output[key] = value.GetValue()
		}
	}
	return output
}
