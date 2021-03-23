package alligotor

import (
	"io"
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
		var fileBytes []byte

		_, err := reader.Read(fileBytes)
		if err != nil {
			return err
		}

		m, err := unmarshal(fileBytes)
		if err != nil {
			continue
		}

		s.fileMaps = append(s.fileMaps, m)
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
