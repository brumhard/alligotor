package alligotor

import (
	"os"
	"path"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("files", func() {
	Describe("unmarshal", func() {
		expectedMap := map[string]interface{}{
			"test": map[string]interface{}{"sub": "lel"},
		}

		Context("yaml", func() {
			It("should succeed with valid input", func() {
				yamlBytes := []byte(`---
test:
  sub: lel
`)
				yamlMap, err := unmarshal(yamlBytes)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(yamlMap.m).To(Equal(expectedMap))
			})
		})
		Context("json", func() {
			It("should succeed with valid input", func() {
				jsonBytes := []byte(`{"test": {"sub": "lel"}}`)
				jsonMap, err := unmarshal(jsonBytes)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(jsonMap.m).To(Equal(expectedMap))
			})
		})
		Context("not supported", func() {
			It("should fail with random input", func() {
				randomBytes := []byte("i don't know what I'm doing here")
				_, err := unmarshal(randomBytes)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(ErrFileTypeNotSupported))
			})
		})
	})
	Describe("readFileMap", func() {
		var (
			m     *ciMap
			field *Field
			name  = "someInt"
		)
		BeforeEach(func() {
			m = newCiMap()
			field = &Field{
				Name:  name,
				value: reflect.ValueOf(0),
			}
		})
		It("returns nil if not set", func() {
			val, err := readFileMap(field, m)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(BeNil())
		})
		It("returns empty string if set to empty string", func() {
			m.m = map[string]interface{}{name: ""}

			val, err := readFileMap(field, m)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal([]byte("")))
		})
		It("return []byte if type mismatch but value is string", func() {
			m.m = map[string]interface{}{name: "1234"}

			val, err := readFileMap(field, m)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal([]byte("1234")))
		})
		It("returns error if type mismatch but value is not a string", func() {
			m.m = map[string]interface{}{name: []string{"1234"}}

			_, err := readFileMap(field, m)
			Expect(err).To(HaveOccurred())
		})
		It("uses configured overwrite long name", func() {
			field.Configs = map[string]string{fileKey: "overwrite"}
			m.m = map[string]interface{}{"overwrite": 3000}

			val, err := readFileMap(field, m)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(3000))
		})
		Context("nested", func() {
			var base = "base"
			BeforeEach(func() {
				field.Base = []string{base}
			})
			It("works", func() {
				m.m = map[string]interface{}{base: map[string]interface{}{name: 1234}}

				val, err := readFileMap(field, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(1234))
			})
			It("name can be overridden, base remains", func() {
				field.Configs = map[string]string{fileKey: "default"}
				m.m = map[string]interface{}{base: map[string]interface{}{"default": 1234}}

				val, err := readFileMap(field, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(1234))
			})
			It("uses overridden name even if normal one is set", func() {
				field.Configs = map[string]string{fileKey: "default"}
				m.m = map[string]interface{}{base: map[string]interface{}{name: 1235, "default": 1234}}

				val, err := readFileMap(field, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(1234))
			})
		})
	})
	Describe("findFiles", func() {
		Context("no location", func() {
			It("returns nil", func() {
				Expect(findFiles(nil, "")).To(BeNil())
			})
		})
		Context("existing location", func() {
			var tmpDir string
			BeforeEach(func() {
				var err error
				tmpDir, err = os.MkdirTemp("", "tests*")
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				Expect(os.RemoveAll(tmpDir)).To(Succeed())
			})
			It("returns nil with no file found", func() {
				Expect(findFiles([]string{tmpDir}, "someFile")).To(BeNil())
			})
			Context("existing file", func() {
				var filePath string
				BeforeEach(func() {
					filePath = path.Join(tmpDir, "test.yml")
					Expect(os.WriteFile(filePath, []byte("test"), os.ModePerm)).To(Succeed())
				})
				It("returns right filePath", func() {
					Expect(findFiles([]string{tmpDir}, "test")).To(Equal([]string{filePath}))
				})
			})
		})
	})
	Describe("FilesSource", func() {
		var s *FilesSource
		Describe("Init", func() {
			BeforeEach(func() {
				var err error
				tmpDir, err := os.MkdirTemp("", "tests*")
				Expect(err).ToNot(HaveOccurred())

				jsonContent := []byte(`{"test":"1234"}`)
				ymlContent := []byte(`test: "1235"`)

				Expect(os.WriteFile(path.Join(tmpDir, "test.json"), jsonContent, os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(path.Join(tmpDir, "test.yml"), ymlContent, os.ModePerm)).To(Succeed())

				s = &FilesSource{
					locations: []string{tmpDir},
					baseName:  "test",
					separator: defaultFileSeparator,
				}
			})
			It("initializes fileMaps", func() {
				Expect(s.Init(nil)).To(Succeed())
				Expect(s.fileMaps).To(Equal([]*ciMap{
					{m: map[string]interface{}{"test": "1234"}},
					{m: map[string]interface{}{"test": "1235"}},
				}))
			})
		})
		Describe("Read", func() {
			var field *Field
			BeforeEach(func() {
				s = &FilesSource{
					fileMaps: []*ciMap{
						{m: map[string]interface{}{"test": "1234"}},
						{m: map[string]interface{}{"test": "1235"}},
					},
					separator: defaultFileSeparator,
				}

				field = &Field{
					Name:  "test",
					value: reflect.ValueOf(""),
				}
			})
			It("fileMaps override each other", func() {
				val, err := s.Read(field)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal("1235"))
			})
		})
	})
})
