package pkg

import (
	"io/ioutil"
	"os"
	"path"
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

		Describe("readEnv", func() {
			var config EnvConfig
			BeforeEach(func() {
				config = EnvConfig{
					Prefix:    "",
					Separator: "_",
					Disabled:  false,
				}
			})
			It("uses uppercase name as default env name", func() {
				err := readEnv(fields, config, map[string]string{"PORT": "3000"})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("uses configured name", func() {
				fields[0].Config.EnvName = "overwrite"
				err := readEnv(fields, config, map[string]string{"OVERWRITE": "3000"})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("uses prefix", func() {
				config.Prefix = "prefix"
				err := readEnv(fields, config, map[string]string{"PREFIX_PORT": "3000"})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("doesn't use prefix if name is configured", func() {
				config.Prefix = "prefix"
				fields[0].Config.EnvName = "overwrite"
				err := readEnv(fields, config, map[string]string{"OVERWRITE": "3000"})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("doesn't overwrite with empty value if not set", func() {
				target.V = 3000
				err := readEnv(fields, config, map[string]string{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(3000))
			})
			It("overwrites with empty value if set to empty", func() {
				target.V = 3000
				err := readEnv(fields, config, map[string]string{"PORT": ""})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(target.V).To(Equal(0))
			})
			Context("nested", func() {
				It("uses separator", func() {
					err := readEnv(nestedFields, config, map[string]string{"SUB_PORT": "1234"})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(nestedTarget.Sub.V).To(Equal(1234))
				})
				It("can be overridden", func() {
					nestedFields[0].Config.EnvName = "PORT"
					err := readEnv(nestedFields, config, map[string]string{"PORT": "1234"})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(nestedTarget.Sub.V).To(Equal(1234))
				})
			})
		})

		Context("files", func() {
			separator := "."

			Describe("readFiles", func() {
				var config FilesConfig
				var baseFileName string
				var dir string
				BeforeEach(func() {
					var err error
					dir, err = ioutil.TempDir("", "tests*")
					Expect(err).ShouldNot(HaveOccurred())

					baseFileName = "testing"
					config = FilesConfig{
						Locations: []string{dir},
						BaseName:  baseFileName,
						Separator: separator,
						Disabled:  false,
					}
				})
				AfterEach(func() {
					Expect(os.RemoveAll(dir)).To(Succeed())
				})
				It("returns an error if no config file is found", func() {
					err := readFiles(fields, config)
					Expect(err).Should(HaveOccurred())
					Expect(err).To(Equal(ErrNoFileFound))
				})
				It("supports yaml, uses name as default file field, ignores extension", func() {
					yamlBytes := []byte(`port: 3000`)
					Expect(ioutil.WriteFile(path.Join(dir, baseFileName+".yaml"), yamlBytes, 0600)).To(Succeed())

					Expect(readFiles(fields, config)).To(Succeed())
					Expect(target.V).To(Equal(3000))
				})
				It("supports json, uses name as default file field, ignores extension", func() {
					jsonBytes := []byte(`{"port":3000}`)
					Expect(ioutil.WriteFile(path.Join(dir, baseFileName), jsonBytes, 0600)).To(Succeed())

					Expect(readFiles(fields, config)).To(Succeed())
					Expect(target.V).To(Equal(3000))
				})
				It("is case insensitive", func() {
					jsonBytes := []byte(`{"PORT":3000}`)
					Expect(ioutil.WriteFile(path.Join(dir, baseFileName), jsonBytes, 0600)).To(Succeed())

					Expect(readFiles(fields, config)).To(Succeed())
					Expect(target.V).To(Equal(3000))
				})
			})

			Describe("readFileMap", func() {
				var m *ciMap
				BeforeEach(func() {
					m = &ciMap{separator: separator}
				})
				It("returns error if field is of wrong type", func() {
					m.m = map[string]interface{}{"port": "1234"}

					Expect(readFileMap(fields, separator, m)).NotTo(Succeed())
				})
				It("uses configured overwrite long name", func() {
					fields[0].Config.FileField = "overwrite"
					m.m = map[string]interface{}{"overwrite": 3000}

					Expect(readFileMap(fields, separator, m)).To(Succeed())
					Expect(target.V).To(Equal(3000))
				})
				It("doesn't overwrite with empty value if not set", func() {
					target.V = 3000

					Expect(readFileMap(fields, separator, m)).To(Succeed())
					Expect(target.V).To(Equal(3000))
				})
				It("overwrites with empty value if set to empty", func() {
					target.V = 3000
					m.m = map[string]interface{}{"port": 0}

					Expect(readFileMap(fields, separator, m)).To(Succeed())
					Expect(target.V).To(Equal(0))
				})
				Context("nested", func() {
					It("works", func() {
						m.m = map[string]interface{}{"sub": map[string]interface{}{"port": 1234}}

						Expect(readFileMap(nestedFields, separator, m)).To(Succeed())
						Expect(nestedTarget.Sub.V).To(Equal(1234))
					})
					It("can be targeted with overwrite", func() {
						nestedFields[0].Config.FileField = "sub.port"
						m.m = map[string]interface{}{"sub": map[string]interface{}{"port": 1234}}

						Expect(readFileMap(nestedFields, separator, m)).To(Succeed())
						Expect(nestedTarget.Sub.V).To(Equal(1234))
					})
					It("can be overridden", func() {
						nestedFields[0].Config.FileField = "over"
						m.m = map[string]interface{}{"over": 1234}

						Expect(readFileMap(nestedFields, separator, m)).To(Succeed())
						Expect(nestedTarget.Sub.V).To(Equal(1234))
					})
				})
			})
		})
	})
	Describe("getEnvAsMap", func() {
		It("gets environment variables in right format", func() {
			Expect(os.Setenv("TESTING_KEY", "TESTING_VAL")).To(Succeed())
			envMap := getEnvAsMap()
			testingVal, ok := envMap["TESTING_KEY"]
			Expect(ok).To(BeTrue())
			Expect(testingVal).To(Equal("TESTING_VAL"))
		})
		It("supports '=' and ',' in the value", func() {
			Expect(os.Setenv("TESTING_KEY", "lel=lol,arr=lul")).To(Succeed())
			envMap := getEnvAsMap()
			testingVal, ok := envMap["TESTING_KEY"]
			Expect(ok).To(BeTrue())
			Expect(testingVal).To(Equal("lel=lol,arr=lul"))
		})
	})
	Describe("readParameterConfig", func() {
		It("returns empty ParameterConfig if configStr is empty", func() {
			p, err := readParameterConfig("")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p).To(Equal(ParameterConfig{}))
		})
		It("panic if configStr hast invalid format", func() {
			Expect(func() { _, _ = readParameterConfig("file=") }).To(Panic())
			Expect(func() { _, _ = readParameterConfig("env") }).To(Panic())
		})
		It("works with valid format configStr, allows whitespace", func() {
			p, err := readParameterConfig("file=val,env=val,flag=l long")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p).To(Equal(ParameterConfig{
				FileField: "val",
				EnvName:   "val",
				Flag: Flag{
					Name:      "long",
					ShortName: "l",
				},
			}))
		})
	})
	Describe("getFieldsConfigsFromValue", func() {
		It("gets correct fields, supports nested struct", func() {
			target := struct {
				Sub struct {
					Port int `config:"env=test"`
				}
			}{}
			fields, err := getFieldsConfigsFromValue(reflect.ValueOf(target))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(fields).To(Equal([]*Field{
				{
					Base:   nil,
					Name:   "Sub",
					Value:  reflect.ValueOf(target.Sub),
					Config: ParameterConfig{},
				},
				{
					Base:  []string{"Sub"},
					Name:  "Port",
					Value: reflect.ValueOf(target.Sub.Port),
					Config: ParameterConfig{
						EnvName: "test",
					},
				},
			}))
		})
	})
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
