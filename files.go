package alligotor

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	fileKey              = "file"
	defaultFileSeparator = "."
)

var ErrFileTypeNotSupported = errors.New("could not unmarshal file, file type not supported")

// FilesConfig is used to configure the configuration from files.
// locations can be used to define where to look for files with the defined baseName.
// Currently only json and yaml files are supported.
// The separator is used for nested structs.
// If Disabled is true the configuration from files is skipped.
type FilesSource struct {
	locations []string
	baseName  string
	separator string
}

// NewFilesSource is a option for New to enable configuration files as configuration source.
// It takes the locations/ dirs where to look for files and the basename (without file extension) as input parameters.
// NewFilesSource itself takes more options to customize the used separator (WithFileSeparator).
func NewFilesSource(locations []string, baseName string, opts ...FileOption) *FilesSource {
	files := &FilesSource{
		locations: locations,
		baseName:  baseName,
		separator: defaultFileSeparator,
	}

	for _, opt := range opts {
		opt(files)
	}

	return files
}

// FileOption takes a FilesConfig as input and modifies it.
type FileOption func(*FilesSource)

// WithFileSeparator adds a custom separator to a FilesConfig struct.
func WithFileSeparator(separator string) FileOption {
	return func(files *FilesSource) {
		files.separator = separator
	}
}

func (f *FilesSource) Read(fields []*Field) error {
	for _, fileLocation := range f.locations {
		fileInfos, err := ioutil.ReadDir(fileLocation)
		if err != nil {
			continue
		}

		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			if strings.TrimSuffix(name, path.Ext(name)) != f.baseName {
				continue
			}

			fileBytes, err := ioutil.ReadFile(path.Join(fileLocation, name))
			if err != nil {
				return err
			}

			m, err := unmarshal(f.separator, fileBytes)
			if err != nil {
				return err
			}

			if err := readFileMap(fields, f.separator, m); err != nil {
				return err
			}
		}
	}

	return nil
}

func readFileMap(fields []*Field, separator string, m *ciMap) error {
	for _, f := range fields {
		fieldNames := []string{
			f.Configs[fileKey],
			f.FullName(separator),
		}

		for _, fieldName := range fieldNames {
			valueForField, ok := m.Get(fieldName)
			if !ok {
				continue
			}

			fieldTypeZero := reflect.Zero(f.Value().Type())
			v := fieldTypeZero.Interface()

			if err := mapstructure.Decode(valueForField, &v); err != nil {
				// if theres a type mismatch check if value is a string and try to use SetFromString (e.g. for duration strings)
				if valueString, ok := valueForField.(string); ok {
					if err := SetFromString(f.Value(), valueString); err != nil {
						return err
					}

					continue
				}

				// if the target is a struct there are also fields for the child properties and it should be tried
				// to set these before returning an error
				if f.Value().Kind() == reflect.Struct {
					continue
				}

				return err
			}

			f.Value().Set(reflect.ValueOf(v))
		}
	}

	return nil
}

func unmarshal(fileSeparator string, bytes []byte) (*ciMap, error) {
	m := newCiMap(withSeparator(fileSeparator))
	if err := yaml.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	if err := json.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	return nil, ErrFileTypeNotSupported
}
