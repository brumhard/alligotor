package alligotor

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type ReadersSource struct {
	readers  []io.Reader
	fileMaps []*ciMap
}

func NewReadersSource(readers ...io.Reader) *ReadersSource {
	return &ReadersSource{
		readers: readers,
	}
}

// Init initializes the fileMaps property.
// It should be used right before calling the Read method to load the latest config files' states.
func (s *ReadersSource) Init(_ []Field) error {
	for _, reader := range s.readers {
		if err := func() error {
			if closer, ok := reader.(io.Closer); ok {
				defer closer.Close()
			}

			m, err := unmarshal(reader)
			if err != nil {
				return nil
			}

			s.fileMaps = append(s.fileMaps, m)

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

// Read reads the saved fileMaps from the Init function and returns the set value for a certain field.
// If not value is set in the flags it returns nil.
func (s *ReadersSource) Read(field *Field) (interface{}, error) {
	var finalVal interface{}

	for _, m := range s.fileMaps {
		val, err := readFileMap(field, m)
		if err != nil {
			return nil, err
		}

		finalVal = val
	}

	return finalVal, nil
}

func unmarshal(r io.Reader) (*ciMap, error) {
	m := newCiMap()
	if err := yaml.NewDecoder(r).Decode(m); err == nil {
		return m, nil
	}

	if err := json.NewDecoder(r).Decode(m); err == nil {
		return m, nil
	}

	return nil, ErrFileTypeNotSupported
}

func readFileMap(f *Field, m *ciMap) (interface{}, error) {
	name := f.Name()
	if f.Configs()[fileKey] != "" {
		name = f.Configs()[fileKey]
	}

	valueForField, ok := m.Get(f.Base(), name)
	if !ok {
		return nil, nil
	}

	fieldTypeNew := reflect.New(f.Type())

	if err := mapstructure.Decode(valueForField, fieldTypeNew.Interface()); err != nil {
		// if theres a type mismatch check if value is a string so maybe it can be parsed
		if valueString, ok := valueForField.(string); ok {
			return []byte(valueString), nil
		}

		// if it's a struct, maybe one of the properties can be assigned nevertheless
		if f.Type().Kind() == reflect.Struct {
			return nil, nil
		}

		return nil, err
	}

	return fieldTypeNew.Elem().Interface(), nil
}
