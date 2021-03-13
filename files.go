package alligotor

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const fileKey = "file"

var ErrFileTypeNotSupported = errors.New("could not unmarshal file, file type not supported or malformed content")

// FilesSource is used to read the configuration from files.
// locations can be used to define where to look for files with the defined baseName.
// Currently only json and yaml files are supported.
// The separator is used for nested structs.
type FilesSource struct {
	locations []string
	baseName  string
	fileMaps  []*ciMap
}

// NewFilesSource returns a new FilesSource.
// It takes the locations/ dirs where to look for files and the basename (without file extension) as input parameters.
// If locations or baseName are empty this is a noop source.
func NewFilesSource(locations []string, baseName string) *FilesSource {
	return &FilesSource{
		locations: locations,
		baseName:  baseName,
	}
}

// Init initializes the fileMaps property.
// It should be used right before calling the Read method to load the latest config files' states.
func (s *FilesSource) Init(_ []*Field) error {
	files := findFiles(s.locations, s.baseName)

	for _, filePath := range files {
		fileBytes, err := os.ReadFile(path.Join(filePath))
		if err != nil {
			continue
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
func (s *FilesSource) Read(f *Field) (interface{}, error) {
	var finalVal interface{}

	for _, m := range s.fileMaps {
		val, err := readFileMap(f, m)
		if err != nil {
			return nil, err
		}

		finalVal = val
	}

	return finalVal, nil
}

func findFiles(locations []string, baseName string) []string {
	if baseName == "" {
		return nil
	}

	var filePaths []string

	for _, fileLocation := range locations {
		fileInfos, err := os.ReadDir(fileLocation)
		if err != nil {
			continue
		}

		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			if strings.TrimSuffix(name, path.Ext(name)) != baseName {
				continue
			}

			filePaths = append(filePaths, path.Join(fileLocation, name))
		}
	}

	return filePaths
}

func unmarshal(bytes []byte) (*ciMap, error) {
	m := newCiMap()
	if err := yaml.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	if err := json.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	return nil, ErrFileTypeNotSupported
}

func readFileMap(f *Field, m *ciMap) (interface{}, error) {
	fieldName := f.Name
	if f.Configs[fileKey] != "" {
		fieldName = f.Configs[fileKey]
	}

	valueForField, ok := m.Get(f.Base, fieldName)
	if !ok {
		return nil, nil
	}

	fieldTypeNew := reflect.New(f.Type())

	if err := mapstructure.Decode(valueForField, fieldTypeNew.Interface()); err != nil {
		// if theres a type mismatch check if value is a string so maybe it can be parsed
		if valueString, ok := valueForField.(string); ok {
			return []byte(valueString), nil
		}

		return nil, err
	}

	return fieldTypeNew.Elem().Interface(), nil
}
