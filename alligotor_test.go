package alligotor

import (
	"os"
	"path"
	"reflect"
	"time"

	"github.com/brumhard/alligotor/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	Describe("specialTypes", func() {
		Context("duration", func() {
			var (
				valString = "2h"
				expected  = 2 * time.Hour
			)
			It("parses normal", func() {
				val, err := specialTypes(reflect.ValueOf(expected), valString)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(expected))
			})
			It("parses pointer", func() {
				val, err := specialTypes(reflect.ValueOf(&expected), valString)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(&expected))
			})
		})
		Context("time", func() {
			var (
				valString = "2007-01-02T15:04:05Z"
				expected  = time.Date(2007, 01, 02, 15, 04, 05, 00, time.UTC)
			)
			It("parses normal", func() {
				val, err := specialTypes(reflect.ValueOf(expected), valString)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(expected))
			})
			It("parses pointer", func() {
				val, err := specialTypes(reflect.ValueOf(&expected), valString)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(&expected))
			})
		})
		Context("string", func() {
			var (
				valString = "test"
				expected  = "test"
			)
			It("parses normal", func() {
				val, err := specialTypes(reflect.ValueOf(expected), valString)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(expected))
			})
			It("parses pointer", func() {
				val, err := specialTypes(reflect.ValueOf(&expected), valString)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(&expected))
			})
		})
		Context("stringSlice", func() {
			var (
				valString = "a,b,c"
				expected  = []string{"a", "b", "c"}
			)
			It("parses normal", func() {
				val, err := specialTypes(reflect.ValueOf(expected), valString)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(expected))
			})
		})
		Context("stringMap", func() {
			var (
				valString = "a=a,b=b,c=c"
				expected  = map[string]string{"a": "a", "b": "b", "c": "c"}
			)
			It("parses normal", func() {
				val, err := specialTypes(reflect.ValueOf(expected), valString)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(expected))
			})
		})
	})
	Describe("fromString", func() {
		It("uses specialTypes for special parsing", func() {
			var (
				valString = "2h"
				expected  = 2 * time.Hour
			)
			val, err := fromString(reflect.ValueOf(expected), valString)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(expected))
		})
		It("supports textunmarshaler", func() {
			var (
				valString = "arr"
				expected  = &testType{S: "arr"}
			)
			val, err := fromString(reflect.ValueOf(expected), valString)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(expected))
		})
		It("uses json.Unmarshal for other types like int", func() {
			var (
				valString = "420"
				expected  = 420
			)
			val, err := fromString(reflect.ValueOf(expected), valString)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(expected))
		})
	})
	Describe("set", func() {
		var target *struct{ V int }
		BeforeEach(func() {
			target = &struct{ V int }{}
		})
		It("uses custom marshaling if value is byte slice", func() {
			Expect(set(wrappedValue(target), []byte("5"))).To(Succeed())
			Expect(target.V).To(Equal(5))
		})
		It("assigns value directly if not byte slice", func() {
			Expect(set(wrappedValue(target), 5)).To(Succeed())
			Expect(target.V).To(Equal(5))
		})
		It("returns error on type mismatch", func() {
			Expect(set(wrappedValue(target), "abc")).To(MatchError(ErrTypeMismatch))
		})
	})
	Describe("readParameterConfig", func() {
		It("returns empty parameterConfig if configStr is empty", func() {
			p, err := readParameterConfig("")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p).To(BeNil())
		})
		It("panic if configStr hast invalid format", func() {
			Expect(func() { _, _ = readParameterConfig("file=") }).To(Panic())
			Expect(func() { _, _ = readParameterConfig("env") }).To(Panic())
		})
		It("works with valid format configStr, allows whitespace", func() {
			p, err := readParameterConfig("file=val,env=val,flag=l long")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p).To(Equal(map[string]string{
				"file": "val",
				"env":  "val",
				"flag": "l long",
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
					Base:    nil,
					Name:    "Sub",
					value:   reflect.ValueOf(target.Sub),
					Configs: nil,
				},
				{
					Base:    []string{"Sub"},
					Name:    "Port",
					value:   reflect.ValueOf(target.Sub.Port),
					Configs: map[string]string{"env": "test"},
				},
			}))
		})
	})
	Describe("Collector", func() {
		Describe("Get", func() {
			var (
				tempDir      string
				c            *Collector
				fileBaseName string
			)

			BeforeEach(func() {
				var err error
				// create temp dir
				tempDir, err = os.MkdirTemp("", "tests*")
				Expect(err).ShouldNot(HaveOccurred())

				fileBaseName = "config"

				c = New(
					NewFilesSource([]string{tempDir}, fileBaseName),
					NewEnvSource(""),
					NewFlagsSource(),
				)
			})
			AfterEach(func() {
				// delete temp dir
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})
			It("returns error if v is not a pointer", func() {
				err := (&Collector{}).Get(struct{}{})
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(ErrPointerExpected))
			})
			It("works if v is a pointer", func() {
				err := (&Collector{}).Get(&struct{}{})
				Expect(err).ShouldNot(HaveOccurred())
			})
			It("supports pointers for properties", func() {
				testingStruct := testingConfigPointers{
					API: &test.APIConfig{Port: 1, LogLevel: "info"},
					DB:  &test.DBConfig{LogLevel: "info"},
				}
				jsonBytes := []byte(`{"logLevel": "default", "api": {"port": 2, "logLevel": "specified"}}`)
				Expect(os.WriteFile(path.Join(tempDir, fileBaseName), jsonBytes, 0600)).To(Succeed())

				Expect(c.Get(&testingStruct)).To(Succeed())
				Expect(testingStruct.API.Port).To(Equal(2))
				Expect(testingStruct.DB.LogLevel).To(Equal("default"))
				Expect(testingStruct.API.LogLevel).To(Equal("specified"))
			})
			It("supports embedded structs for properties", func() {
				testingStruct := testingConfigEmbedded{
					APIConfig: test.APIConfig{Port: 1, LogLevel: "info"},
					DBConfig:  test.DBConfig{LogLevel: "info"},
				}
				jsonBytes := []byte(`{"logLevel": "default", "apiConfig": {"port": 2, "logLevel": "specified"}}`)
				Expect(os.WriteFile(path.Join(tempDir, fileBaseName), jsonBytes, 0600)).To(Succeed())

				Expect(c.Get(&testingStruct)).To(Succeed())
				Expect(testingStruct.APIConfig.Port).To(Equal(2))
				Expect(testingStruct.DBConfig.LogLevel).To(Equal("default"))
				Expect(testingStruct.APIConfig.LogLevel).To(Equal("specified"))
			})
			Context("Integration Tests", func() {
				var args []string
				var env map[string]string
				BeforeEach(func() {
					// capture os.Args
					args = os.Args
					// capture env
					env = getEnvAsMap()
				})
				AfterEach(func() {
					// recover os.Args
					os.Args = args
					// recover env
					for k, v := range env {
						Expect(os.Setenv(k, v)).To(Succeed())
					}
				})
				Describe("overwrites: default, file, env, flag", func() {
					testingStruct := testingConfig{
						Enabled: false,
						Sleep:   time.Minute,
						API:     test.APIConfig{Port: 1, LogLevel: "info"},
						DB:      test.DBConfig{LogLevel: "info"},
					}
					defaults := testingStruct

					It("keeps defaults if no values are set", func() {
						Expect(c.Get(&testingStruct)).To(Succeed())
						Expect(testingStruct).To(Equal(defaults))
					})
					Context("file is set", func() {
						BeforeEach(func() {
							jsonBytes := []byte(`{"logLevel": "default", "sleep": "1s", "api": {"port": 2, "logLevel": "specifiedInFile"}}`)
							Expect(os.WriteFile(path.Join(tempDir, fileBaseName), jsonBytes, 0600)).To(Succeed())
						})
						It("overrides defaults", func() {
							Expect(c.Get(&testingStruct)).To(Succeed())
							Expect(testingStruct.Sleep).To(Equal(1 * time.Second))
							Expect(testingStruct.API.Port).To(Equal(2))
							Expect(testingStruct.DB.LogLevel).To(Equal("default"))
							Expect(testingStruct.API.LogLevel).To(Equal("specifiedInFile"))
						})
						Context("env is set", func() {
							BeforeEach(func() {
								Expect(os.Setenv("PORT", "3")).To(Succeed())
								Expect(os.Setenv("DB_LOGLEVEL", "logLevelFromEnv")).To(Succeed())
								Expect(os.Setenv("SLEEP", "2m")).To(Succeed())
							})
							It("overrides file", func() {
								Expect(c.Get(&testingStruct)).To(Succeed())
								Expect(testingStruct.Sleep).To(Equal(2 * time.Minute))
								Expect(testingStruct.API.Port).To(Equal(3))
								Expect(testingStruct.DB.LogLevel).To(Equal("logLevelFromEnv"))
								Expect(testingStruct.API.LogLevel).To(Equal("specifiedInFile"))
							})
							Context("flags are set", func() {
								BeforeEach(func() {
									os.Args = []string{"commandName", "-p", "4", "--enabled", "true", "--sleep", "3h"}
								})
								It("overrides env", func() {
									Expect(c.Get(&testingStruct)).To(Succeed())
									Expect(testingStruct.Enabled).To(Equal(true))
									Expect(testingStruct.Sleep).To(Equal(3 * time.Hour))
									Expect(testingStruct.API.Port).To(Equal(4))
									Expect(testingStruct.DB.LogLevel).To(Equal("logLevelFromEnv"))
									Expect(testingStruct.API.LogLevel).To(Equal("specifiedInFile"))
								})
							})
						})
					})
				})
			})
		})
		Describe("custom string map and slice", func() {
			Describe("string slice", func() {
				It("works for Unmarshal", func() {
					s := stringSlice{}
					Expect(s.UnmarshalText([]byte("string, lol, lel"))).To(Succeed())
					Expect([]string(s)).To(Equal([]string{"string", "lol", "lel"}))
				})
			})
			Describe("string map", func() {
				It("works for Unmarshal", func() {
					s := stringMap{}
					Expect(s.UnmarshalText([]byte("field1 = string, field2 = lol"))).To(Succeed())
					Expect(map[string]string(s)).To(Equal(map[string]string{
						"field1": "string",
						"field2": "lol",
					}))
				})
			})
		})
	})
})

type testingConfig struct {
	Enabled bool
	Sleep   time.Duration
	API     test.APIConfig
	DB      test.DBConfig
}

type testingConfigPointers struct {
	API *test.APIConfig
	DB  *test.DBConfig
}

type testingConfigEmbedded struct {
	test.APIConfig
	test.DBConfig
}

func wrappedValue(val interface{}) reflect.Value {
	return reflect.ValueOf(val).Elem().Field(0)
}

type testType struct {
	S string
}

func (t *testType) UnmarshalText(text []byte) error {
	t.S = string(text)

	return nil
}
