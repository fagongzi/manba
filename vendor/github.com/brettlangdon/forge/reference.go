package forge

import "errors"

// Reference struct used for holding data neede for Reference data type
type Reference struct {
	section *Section
	name    string
}

// NewReference will create and initialize a new Reference value
func NewReference(name string, section *Section) *Reference {
	return &Reference{
		section: section,
		name:    name,
	}
}

func (reference *Reference) resolve() Value {
	value, err := reference.section.Resolve(reference.name)
	if err != nil {
		value = NewNull()
	}
	return value
}

// GetType will simply return back REFERENCE
func (reference *Reference) GetType() ValueType {
	return REFERENCE
}

// GetValue will resolve and return the value from the underlying reference
func (reference *Reference) GetValue() interface{} {
	return reference.resolve().GetValue()
}

// UpdateValue will simply throw an error since it is not allowed for References
func (reference *Reference) UpdateValue(value interface{}) error {
	return errors.New("cannot update value of a reference")
}
