package alligotor

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

const fileKey = "file"

var ErrFileFormatNotSupported = errors.New("file format not supported or malformed content")

// ReadersSource is used to read configuration from any type that implements the io.Reader interface.
// The data in the readers should be in one of the supported file formats (currently yml and json).
// This enables a wide range of usages like for example reading the config from an http endpoint or a file.
//
// The ReadersSource accepts io.Reader to support as many types as possible. To improve the experience with sources
// that need to be closed it will also check if the supplied type implements io.Closer and closes the reader
// if it does.
type ReadersSource struct {
	readers  []io.Reader
	fileMaps []*ciMap
}

// NewReadersSource returns a new ReadersSource that reads from one or more readers.
// If the input reader slice is empty this will be a noop reader.
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

// unmarshal tries to decode the reader's data into any supported fileType. If it does not work for any file format
// an ErrFileFormatNotSupported is returned.
func unmarshal(r io.Reader) (*ciMap, error) {
	m := newCiMap()
	if err := yaml.NewDecoder(r).Decode(m); err == nil {
		return m, nil
	}

	if err := json.NewDecoder(r).Decode(m); err == nil {
		return m, nil
	}

	return nil, ErrFileFormatNotSupported
}

// readFileMap reads the value for a given field from the given ciMap.
// It returns the right type if there is no decoding error otherwise it returns a byte slice that could potentially
// be decoded later into the target type.
func readFileMap(f *Field, m *ciMap) (interface{}, error) {
	name := extractFileName(f)

	valueForField, ok := m.Get(f.BaseNames(extractFileName), name)
	if !ok {
		return nil, nil
	}

	fieldTypeNew := reflect.New(f.Type())

	if f.Type().Kind() == reflect.Struct {
		// if it's a struct, it could be assigned with TextUnmarshaler, otherwise return nil
		if valueString, ok := valueForField.(string); ok {
			return []byte(valueString), nil
		}

		return nil, nil
	}

	if err := mapstructure.Decode(valueForField, fieldTypeNew.Interface()); err != nil {
		// if theres a type mismatch check if value is a string so maybe it can be parsed
		if valueString, ok := valueForField.(string); ok {
			return []byte(valueString), nil
		}

		return nil, err
	}

	return fieldTypeNew.Elem().Interface(), nil
}

func extractFileName(f *Field) string {
	if f.Configs()[fileKey] != "" {
		return f.Configs()[fileKey]
	}

	return f.Name()
}
