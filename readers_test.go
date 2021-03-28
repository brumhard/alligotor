package alligotor

import (
	"bytes"
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
				yamlMap, err := unmarshal(bytes.NewReader(yamlBytes))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(yamlMap.m).To(Equal(expectedMap))
			})
		})
		Context("json", func() {
			It("should succeed with valid input", func() {
				jsonBytes := []byte(`{"test": {"sub": "lel"}}`)
				jsonMap, err := unmarshal(bytes.NewReader(jsonBytes))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(jsonMap.m).To(Equal(expectedMap))
			})
		})
		Context("not supported", func() {
			It("should fail with random input", func() {
				randomBytes := []byte("i don't know what I'm doing here")
				_, err := unmarshal(bytes.NewReader(randomBytes))
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(ErrFileFormatNotSupported))
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
				name:  name,
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
			field.configs = map[string]string{fileKey: "overwrite"}
			m.m = map[string]interface{}{"overwrite": 3000}

			val, err := readFileMap(field, m)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(3000))
		})
		Context("nested", func() {
			var base = "base"
			BeforeEach(func() {
				field.base = []string{base}
			})
			It("works", func() {
				m.m = map[string]interface{}{base: map[string]interface{}{name: 1234}}

				val, err := readFileMap(field, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(1234))
			})
			It("name can be overridden, base remains", func() {
				field.configs = map[string]string{fileKey: "default"}
				m.m = map[string]interface{}{base: map[string]interface{}{"default": 1234}}

				val, err := readFileMap(field, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(1234))
			})
			It("uses overridden name even if normal one is set", func() {
				field.configs = map[string]string{fileKey: "default"}
				m.m = map[string]interface{}{base: map[string]interface{}{name: 1235, "default": 1234}}

				val, err := readFileMap(field, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(1234))
			})
		})
	})
	Describe("ReadersSource", func() {
		var s *ReadersSource
		Describe("Init", func() {
			BeforeEach(func() {
				jsonContent := []byte(`{"test":"1234"}`)
				ymlContent := []byte(`test: "1235"`)

				s = NewReadersSource(bytes.NewReader(jsonContent), bytes.NewReader(ymlContent))
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
				s = &ReadersSource{
					fileMaps: []*ciMap{
						{m: map[string]interface{}{"test": "1234"}},
						{m: map[string]interface{}{"test": "1235"}},
					},
				}

				field = &Field{
					name:  "test",
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
