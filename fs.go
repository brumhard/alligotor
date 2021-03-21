package alligotor

import (
	"io/fs"
	"os"
)

// FilesSource is used to read the configuration from files.
// locations can be used to define where to look for files with the defined baseNames.
// The baseNames define the names of the files to look for without the file extension
// (as multiple file types are supported).
// Currently only json and yaml files are supported.
type FSSource struct {
	fsys     fs.FS
	globs    []string
	fileMaps []*ciMap
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
	return NewFSSource(os.DirFS(""), globs...)
}

// Init initializes the fileMaps property.
// It should be used right before calling the Read method to load the latest config files' states.
func (s *FSSource) Init(_ []Field) error {
	files, err := findFSFiles(s.fsys, s.globs)
	if err != nil {
		return err
	}

	for _, filePath := range files {
		fileBytes, err := fs.ReadFile(s.fsys, filePath)
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
func (s *FSSource) Read(field *Field) (interface{}, error) {
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

func findFSFiles(fsys fs.FS, globs []string) ([]string, error) {
	var filePaths []string

	for _, glob := range globs {
		matches, err := fs.Glob(fsys, glob)
		if err != nil {
			return nil, err
		}

		filePaths = append(filePaths, matches...)
	}

	return filePaths, nil
}
