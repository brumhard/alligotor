package alligotor

import (
	"reflect"
	"strings"
)

type Field struct {
	base  []string
	name  string
	value reflect.Value
	// Configs contains structtag key -> value string
	Configs map[string]string
}

func (f *Field) FullName(separator string) string {
	return strings.Join(append(f.base, f.name), separator)
}

func (f *Field) Value() reflect.Value {
	return f.value
}

type ConfigSource interface {
	Read(fields []*Field) error
}
