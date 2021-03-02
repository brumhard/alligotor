package alligotor

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"
	"sync"

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
	// lazy loading
	once     sync.Once
	fileMaps []*ciMap
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

func (s *FilesSource) Read(f *Field) ([]byte, error) {
	if s.fileMaps == nil {
		s.once.Do(s.setup)
	}

	var finalVal []byte
	for _, m := range s.fileMaps {
		val, err := readFileMap(f, m, s.separator)
		if err != nil {
			return nil, err
		}

		finalVal = val
	}

	return finalVal, nil
}

func (s *FilesSource) setup() {
	files := findFiles(s.locations, s.baseName)

	for _, filePath := range files {
		fileBytes, err := ioutil.ReadFile(path.Join(filePath))
		if err != nil {
			continue
		}

		m, err := unmarshal(fileBytes, s.separator)
		if err != nil {
			continue
		}

		s.fileMaps = append(s.fileMaps, m)
	}
}

func findFiles(locations []string, baseName string) []string {
	var filePaths []string
	for _, fileLocation := range locations {
		fileInfos, err := ioutil.ReadDir(fileLocation)
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

func unmarshal(bytes []byte, fileSeparator string) (*ciMap, error) {
	m := newCiMap(withSeparator(fileSeparator))
	if err := yaml.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	if err := json.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	return nil, ErrFileTypeNotSupported
}

func readFileMap(f *Field, m *ciMap, separator string) ([]byte, error) {
	fieldNames := []string{
		f.Configs[fileKey],
		f.FullName(separator),
	}

	var finalVal []byte
	for _, fieldName := range fieldNames {
		valueForField, ok := m.Get(fieldName)
		if !ok {
			continue
		}

		valueBytes, err := json.Marshal(valueForField)
		if err != nil {
			return nil, err
		}

		finalVal = valueBytes
	}

	return finalVal, nil
}