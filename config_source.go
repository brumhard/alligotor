package alligotor

import (
	"reflect"
)

// Field is a struct to hold all information for a struct's field that should be filled with configuration.
type Field struct {
	// base contains all the parents properties' names in order.
	base []string
	// name is the name of the current property.
	name string
	// description contains the value contained in the description struct tag.
	description string
	// value contains the reflect.Value of the field to set it's value.
	value reflect.Value
	// configs contains structtag key -> value string and can be read to interpret the field's struct tags for
	// custom behavior like overrides.
	configs map[string]string
}

func NewField(base []string, name, description string, value reflect.Value, configs map[string]string) Field {
	return Field{
		base:        base,
		name:        name,
		description: description,
		value:       value,
		configs:     configs,
	}
}

func (f *Field) Base() []string {
	return f.base
}

func (f *Field) Name() string {
	return f.name
}

func (f *Field) Description() string {
	return f.description
}

func (f *Field) Configs() map[string]string {
	return f.configs
}

// Type returns the type of the package. This can be used to switch on the type to parse for example a string
// to the right target type.
func (f *Field) Type() reflect.Type {
	return f.value.Type()
}

// ConfigSource consists of one method that gets a certain field and should return its value.
// If this value is a string and should be parsed (for example env variables can only be retrieved as a string but
// could also resemble an int value or even a string slice), a []byte should be returned.
//
// If anything else than a byte slice is returned the given value will be used as is and if there's a type mismatch
// an error will be reported.
type ConfigSource interface {
	Read(field *Field) (interface{}, error)
}

// ConfigSourceInitializer is an optional interface to implement and can be used to initialize the config source
// before reading the fields one by one with the Read method of ConfigSource.
type ConfigSourceInitializer interface {
	// Init should be called right before Read to initialize stuff.
	// Some things shouldn't be initialized in the constructor since the environment or files (the config source)
	// could be altered in the time between constructing a config source and calling the Read method.
	Init(fields []Field) error
}
