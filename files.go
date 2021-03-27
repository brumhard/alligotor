package alligotor

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

const fileKey = "file"

var ErrFileTypeNotSupported = errors.New("could not unmarshal file, file type not supported or malformed content")

// FilesSource is used to read the configuration from files.
// locations can be used to define where to look for files with the defined baseNames.
// The baseNames define the names of the files to look for without the file extension
// (as multiple file types are supported).
// Currently only json and yaml files are supported.
type FilesSource struct {
	globs []string
	ReadersSource
}

// NewFilesSource returns a new FilesSource.
// It takes the locations/ dirs where to look for files and the baseNames (without file extension) as input parameters.
// If locations or baseNames are empty this is a noop source.
func NewFilesSource(globs []string) *FilesSource {
	return &FilesSource{
		globs: globs,
	}
}

// Init initializes the fileMaps property.
// It should be used right before calling the Read method to load the latest config files' states.
func (s *FilesSource) Init(fields []Field) error {
	files, err := loadOSFiles(s.globs)
	if err != nil {
		return err
	}

	s.ReadersSource = *NewReadersSource(files...)

	return s.ReadersSource.Init(fields)
}

func loadOSFiles(globs []string) ([]io.Reader, error) {
	var files []io.Reader

	for _, glob := range globs {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			file, err := os.Open(match)
			if err != nil {
				return nil, err
			}

			files = append(files, file)
		}
	}

	return files, nil
}
