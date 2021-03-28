package alligotor

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type (
	globFunc = func(pattern string) ([]string, error)
	openFunc = func(path string) (io.Reader, error)
)

// FilesSource is a wrapper around ReadersSource to automatically find the needed readers in a filesystem.
// To use the local FS the NewFilesSource can be used to find the files.
// NewFSFilesSource can be used for other usecases where the FS is present on S3 or an embed.FS.
// This differentiation was implemented since the os.DirFS does not really support relative and absolute paths
// easily.
type FilesSource struct {
	globs []string
	// globFunc is used to match globs on the selected file system
	// The to used implementations are fs.Glob and filepath.Glob
	globFunc globFunc
	// openFunc is used to open a file in the selected file system
	// The used implementations are fsys.Open and os.Open
	openFunc openFunc
	ReadersSource
}

// Init tries to find files on the filesystem matching the supplied globs and reads them.
// Afterwards the underlying ReadersSource is initialized.
func (s *FilesSource) Init(fields []Field) error {
	files, err := loadFiles(s.globs, s.globFunc, s.openFunc)
	if err != nil {
		return err
	}

	s.ReadersSource = *NewReadersSource(files...)

	return s.ReadersSource.Init(fields)
}

// loadFiles tries to find files that match the globs using the globF function.
// If any matches are found it then opens the file using the openF function and returns the opened files.
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
