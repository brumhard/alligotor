package alligotor

import (
	"reflect"
	"strings"
)

type Field struct {
	Base  []string
	Name  string
	value reflect.Value
	// Configs contains structtag key -> value string
	Configs map[string]string
}

func (f *Field) FullName(separator string) string {
	return strings.Join(append(f.Base, f.Name), separator)
}

func (f *Field) Type() reflect.Type {
	return f.value.Type()
}

type ConfigSource interface {
	Read(field *Field) (interface{}, error)
}

type ConfigSourceInitializer interface {
	// Init should be called right before Read to initialize stuff.
	// Some things shouldn't be initialized in the constructor since the environment or files (the config source)
	// could be altered in the time between constructing a config source and calling the Read method.
	Init(fields []*Field) error
}
