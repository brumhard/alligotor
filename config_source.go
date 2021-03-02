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

type ConfigSource interface {
	Read(field *Field) ([]byte, error)
}
