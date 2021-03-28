package alligotor

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const fileKey = "file"

var ErrFileTypeNotSupported = errors.New("could not unmarshal file, file type not supported or malformed content")

type globFunc func(pattern string) ([]string, error)

type openFunc func(path string) (io.Reader, error)

// FilesSource is used to read the configuration from files.
// locations can be used to define where to look for files with the defined baseNames.
// The baseNames define the names of the files to look for without the file extension
// (as multiple file types are supported).
// Currently only json and yaml files are supported.
type FilesSource struct {
	globs    []string
	globFunc globFunc
	openFunc openFunc
	ReadersSource
}

// Init initializes the fileMaps property.
// It should be used right before calling the Read method to load the latest config files' states.
func (s *FilesSource) Init(fields []Field) error {
	files, err := loadFiles(s.globs, s.globFunc, s.openFunc)
	if err != nil {
		return err
	}

	s.ReadersSource = *NewReadersSource(files...)

	return s.ReadersSource.Init(fields)
}

func loadFiles(globs []string, globF globFunc, openF openFunc) ([]io.Reader, error) {
	var files []io.Reader

	for _, glob := range globs {
		matches, err := globF(glob)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			file, err := openF(match)
			if err != nil {
				return nil, err
			}

			files = append(files, file)
		}
	}

	return files, nil
}

func NewFSFilesSource(fsys fs.FS, globs ...string) *FilesSource {
	return &FilesSource{
		globs: globs,
		globFunc: func(pattern string) ([]string, error) {
			return fs.Glob(fsys, pattern)
		},
		openFunc: func(path string) (io.Reader, error) {
			return fsys.Open(path)
		},
	}
}

func NewFilesSource(globs ...string) *FilesSource {
	return &FilesSource{
		globs:    globs,
		globFunc: filepath.Glob,
		openFunc: func(path string) (io.Reader, error) {
			return os.Open(path)
		},
	}
}
