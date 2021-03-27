package alligotor

import (
	"io"
	"io/fs"
	"os"
)

type FSSource struct {
	fsys  fs.FS
	globs []string
	ReadersSource
}

func NewFSSource(fsys fs.FS, globs ...string) *FSSource {
	return &FSSource{
		fsys:  fsys,
		globs: globs,
	}
}

func NewOSFSSource(globs ...string) *FSSource {
	// TODO: does this support relative paths like ../../test.*?
	// TODO: this does not allow absolute paths as well, needs to be fixed
	// 			- https://github.com/golang/go/issues/44286
	// 			- https://github.com/golang/go/issues/44279
	return NewFSSource(os.DirFS(""), globs...)
}

// Init initializes the fileMaps property.
// It should be used right before calling the Read method to load the latest config files' states.
func (s *FSSource) Init(fields []Field) error {
	files, err := loadFSFiles(s.fsys, s.globs)
	if err != nil {
		return err
	}

	s.ReadersSource = *NewReadersSource(files...)

	return s.ReadersSource.Init(fields)
}

func loadFSFiles(fsys fs.FS, globs []string) ([]io.Reader, error) {
	var filesBytes []io.Reader

	for _, glob := range globs {
		matches, err := fs.Glob(fsys, glob)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			file, err := fsys.Open(match)
			if err != nil {
				return nil, err
			}

			filesBytes = append(filesBytes, file)
		}
	}

	return filesBytes, nil
}
