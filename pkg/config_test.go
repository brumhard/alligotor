package pkg

import (
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	Describe("readFlagConfig", func() {
		Context("invalid input", func() {
			It("should return err on more than 3 sets", func() {
				_, err := readFlagConfig("a b c")
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(ErrMalformedFlagConfig))
			})
			It("should return error if longname has less than 2 letters", func() {
				for _, configStr := range []string{"a b", "long long"} {
					_, err := readFlagConfig(configStr)
					Expect(err).Should(HaveOccurred())
					Expect(err).To(Equal(ErrMalformedFlagConfig))
				}
			})
		})
		Context("valid input", func() {
			It("should return valid flag when short and long are set", func() {
				for _, configStr := range []string{"a awd", "awd a"} {
					flag, err := readFlagConfig(configStr)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(flag).To(Equal(Flag{ShortName: "a", Name: "awd"}))
				}
			})
			It("should return valid flag when only short is set", func() {
				flag, err := readFlagConfig("a")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(flag).To(Equal(Flag{ShortName: "a", Name: ""}))
			})
			It("should return valid flag when only long is set", func() {
				flag, err := readFlagConfig("awd")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(flag).To(Equal(Flag{ShortName: "", Name: "awd"}))
			})
		})
	})

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
				yamlMap, err := unmarshal(defaultFileSeparator, yamlBytes)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(yamlMap.m).To(Equal(expectedMap))
			})
		})
		Context("json", func() {
			It("should succeed with valid input", func() {
				jsonBytes := []byte(`{"test": {"sub": "lel"}}`)
				jsonMap, err := unmarshal(defaultFileSeparator, jsonBytes)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(jsonMap.m).To(Equal(expectedMap))
			})
		})
		Context("not supported", func() {
			It("should fail with random input", func() {
				randomBytes := []byte("i don't know what I'm doing here")
				_, err := unmarshal(defaultFileSeparator, randomBytes)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(ErrFileTypeNotSupported))
			})
		})
	})

	Describe("setFromString", func() {
		It("sets anything to zero value if input is empty string", func() {
			target := &struct{ V testType }{testType{S: "testing"}}
			Expect(setFromString(wrappedValue(target), "")).To(Succeed())
			Expect(target.V).To(Equal(testType{}))
		})
		It("sets durations correctly", func() {
			target := &struct{ V time.Duration }{}
			Expect(setFromString(wrappedValue(target), "2s")).To(Succeed())
			Expect(target.V).To(Equal(2 * time.Second))
		})
		It("sets dates correctly", func() {
			target := &struct{ V time.Time }{}
			Expect(setFromString(wrappedValue(target), "2007-01-02T15:04:05Z")).To(Succeed())
			Expect(target.V).To(BeEquivalentTo(time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC)))
		})
		It("sets int types correctly", func() {
			target := &struct{ V int }{}
			Expect(setFromString(wrappedValue(target), "69")).To(Succeed())
			Expect(target.V).To(Equal(69))
		})
		It("sets booleans correctly", func() {
			target := &struct{ V bool }{}
			Expect(setFromString(wrappedValue(target), "true")).To(Succeed())
			Expect(target.V).To(Equal(true))
		})
		It("sets complex types correctly", func() {
			target := &struct{ V complex128 }{}
			Expect(setFromString(wrappedValue(target), "2+3i")).To(Succeed())
			Expect(target.V).To(Equal(complex(2, 3)))
		})
		It("sets uint types correctly", func() {
			target := &struct{ V uint }{}
			Expect(setFromString(wrappedValue(target), "420")).To(Succeed())
			Expect(target.V).To(Equal(uint(420)))
		})
		It("sets float types correctly", func() {
			target := &struct{ V float64 }{}
			Expect(setFromString(wrappedValue(target), "2.34")).To(Succeed())
			Expect(target.V).To(Equal(2.34))
		})
		It("sets strings correctly", func() {
			target := &struct{ V string }{}
			Expect(setFromString(wrappedValue(target), "whoop")).To(Succeed())
			Expect(target.V).To(Equal("whoop"))
		})
		It("sets []string correctly", func() {
			target := &struct{ V []string }{}
			Expect(setFromString(wrappedValue(target), "wow,insane")).To(Succeed())
			Expect(target.V).To(Equal([]string{"wow", "insane"}))
		})
		It("sets map[string]string correctly", func() {
			target := &struct{ V map[string]string }{}
			Expect(setFromString(wrappedValue(target), "wow=insane")).To(Succeed())
			Expect(target.V).To(Equal(map[string]string{"wow": "insane"}))
		})
		It("sets TextUnmarshaler correctly", func() {
			target := &struct{ V testType }{}
			Expect(setFromString(wrappedValue(target), "mmh")).To(Succeed())
			Expect(target.V).To(Equal(testType{S: "mmh"}))
		})
	})

	Context("field function", func() {
		var target *struct{ V int }
		var fields []*Field
		var nestedTarget *struct{ Sub *struct{ V int } }
		var nestedFields []*Field

		BeforeEach(func() {
			target = &struct{ V int }{}
			fields = []*Field{
				{
					Name:   "port",
					Value:  wrappedValue(target),
					Config: ParameterConfig{},
				},
			}

			nestedTarget = &struct{ Sub *struct{ V int } }{
				Sub: &struct{ V int }{V: 0},
			}
			nestedFields = []*Field{
				{
					Base:   []string{"sub"},
					Name:   "port",
					Value:  wrappedValue(nestedTarget, withNested()),
					Config: ParameterConfig{},
				},
			}
		})

		Describe("readPFlags", func() {
			config := FlagsConfig{
				Separator: "-",
				Disabled:  false,
			}

			It("uses name as default flag name", func() {
				err := readPFlags(fields, config, []string{"--port", "3000"})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("uses configured long name", func() {
				fields[0].Config.Flag.Name = "overwrite"
				err := readPFlags(fields, config, []string{"--overwrite", "3000"})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("uses configured short name", func() {
				fields[0].Config.Flag.ShortName = "o"
				err := readPFlags(fields, config, []string{"-o", "3000"})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("doesn't overwrite with empty value if not set", func() {
				target.V = 3000
				err := readPFlags(fields, config, []string{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("overwrites with empty value if set to empty", func() {
				target.V = 3000
				err := readPFlags(fields, config, []string{"--port", ""})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(0))
			})
			Context("nested", func() {
				It("uses separator", func() {
					err := readPFlags(nestedFields, config, []string{"--sub-port", "1234"})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(nestedTarget.Sub.V).To(Equal(1234))
				})
				It("can be overridden", func() {
					nestedFields[0].Config.Flag.Name = "over"
					err := readPFlags(nestedFields, config, []string{"--over", "1234"})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(nestedTarget.Sub.V).To(Equal(1234))
				})
			})
		})

		Describe("readEnv", func() {})
		Describe("readFiles", func() {})
	})
	Describe("getEnvAsMap", func() {})
	Describe("readFieldConfig", func() {})
	Describe("getFieldsConfigsFromValue", func() {})
	Describe("Collector", func() {
		Describe("Get", func() {})
	})
})

type wrapSettings struct {
	nested bool
}

type wrapOption func(options *wrapSettings)

func withNested() wrapOption {
	return func(o *wrapSettings) {
		o.nested = true
	}
}

func wrappedValue(val interface{}, opts ...wrapOption) reflect.Value {
	settings := &wrapSettings{}
	for _, opt := range opts {
		opt(settings)
	}
	if settings.nested {
		return reflect.ValueOf(val).Elem().Field(0).Elem().Field(0)
	}
	return reflect.ValueOf(val).Elem().Field(0)
}

type testType struct {
	S string
}

func (t *testType) UnmarshalText(text []byte) error {
	t.S = string(text)

	return nil
}
