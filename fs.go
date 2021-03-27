package alligotor

import (
	"io"
	"io/fs"
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
	var files []io.Reader

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

			files = append(files, file)
		}
	}

	return files, nil
}
