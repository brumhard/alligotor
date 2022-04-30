package alligotor

import (
	"io"
	"os"
	"path"
	"strings"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("files", func() {
	Describe("loadFiles", func() {
		var (
			nilGlobF = func(_ string) ([]string, error) {
				return nil, nil
			}
			globF = func(pattern string) ([]string, error) {
				return []string{pattern}, nil
			}
			openF = func(path string) (io.Reader, error) {
				return strings.NewReader(path), nil
			}
		)
		Context("no globs", func() {
			It("returns empty slice", func() {
				readers, err := loadFiles(nil, globF, openF)
				Expect(err).ToNot(HaveOccurred())
				Expect(readers).To(HaveLen(0))
			})
		})
		Context("no matches found for globs", func() {
			It("returns empty slice", func() {
				readers, err := loadFiles([]string{"test1", "test2"}, nilGlobF, openF)
				Expect(err).ToNot(HaveOccurred())
				Expect(readers).To(HaveLen(0))
			})
		})
		Context("existing matches", func() {
			It("returns readers for all matches", func() {
				input := []string{"test1", "test2"}

				readers, err := loadFiles(input, globF, openF)
				Expect(err).ToNot(HaveOccurred())
				Expect(readers).To(HaveLen(2))

				var contents []string
				for _, reader := range readers {
					content, err := io.ReadAll(reader)
					Expect(err).NotTo(HaveOccurred())
					contents = append(contents, string(content))
				}
				Expect(input).To(Equal(contents))
			})
		})
	})
	Describe("NewFSFilesSource", func() {
		var s *FilesSource
		BeforeEach(func() {
			s = NewFSFilesSource(
				fstest.MapFS(map[string]*fstest.MapFile{
					"test.json": {Data: []byte(`{"test":"json"}`)},
					"test.yml":  {Data: []byte("test: yml")},
				}),
				"test.*",
			)
		})
		It("loads the files fileMaps correctly", func() {
			Expect(s.Init(nil)).To(Succeed())
			Expect(s.fileMaps).To(Equal([]*ciMap{
				{m: map[string]interface{}{"test": "json"}},
				{m: map[string]interface{}{"test": "yml"}},
			}))
		})
	})
	Describe("NewFilesSource", func() {
		var (
			s      *FilesSource
			tmpDir string
		)
		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "tests*")
			Expect(err).ToNot(HaveOccurred())

			jsonContent := []byte(`{"test":"json"}`)
			ymlContent := []byte(`test: "yml"`)

			Expect(os.WriteFile(path.Join(tmpDir, "test.json"), jsonContent, os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(path.Join(tmpDir, "test.yml"), ymlContent, os.ModePerm)).To(Succeed())

			s = NewFilesSource(path.Join(tmpDir, "test.*"))
		})
		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})
		It("loads the files fileMaps correctly", func() {
			Expect(s.Init(nil)).To(Succeed())
			Expect(s.fileMaps).To(Equal([]*ciMap{
				{m: map[string]interface{}{"test": "json"}},
				{m: map[string]interface{}{"test": "yml"}},
			}))
		})
	})
})
