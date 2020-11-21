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
			Expect(setFromString(WrappedValue(target), "")).To(Succeed())
			Expect(target.V).To(Equal(testType{}))
		})
		It("sets durations correctly", func() {
			target := &struct{ V time.Duration }{}
			Expect(setFromString(WrappedValue(target), "2s")).To(Succeed())
			Expect(target.V).To(Equal(2 * time.Second))
		})
		It("sets dates correctly", func() {
			target := &struct{ V time.Time }{}
			Expect(setFromString(WrappedValue(target), "2007-01-02T15:04:05Z")).To(Succeed())
			Expect(target.V).To(BeEquivalentTo(time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC)))
		})
		It("sets int types correctly", func() {
			target := &struct{ V int }{}
			Expect(setFromString(WrappedValue(target), "69")).To(Succeed())
			Expect(target.V).To(Equal(69))
		})
		It("sets booleans correctly", func() {
			target := &struct{ V bool }{}
			Expect(setFromString(WrappedValue(target), "true")).To(Succeed())
			Expect(target.V).To(Equal(true))
		})
		It("sets complex types correctly", func() {
			target := &struct{ V complex128 }{}
			Expect(setFromString(WrappedValue(target), "2+3i")).To(Succeed())
			Expect(target.V).To(Equal(complex(2, 3)))
		})
		It("sets uint types correctly", func() {
			target := &struct{ V uint }{}
			Expect(setFromString(WrappedValue(target), "420")).To(Succeed())
			Expect(target.V).To(Equal(uint(420)))
		})
		It("sets float types correctly", func() {
			target := &struct{ V float64 }{}
			Expect(setFromString(WrappedValue(target), "2.34")).To(Succeed())
			Expect(target.V).To(Equal(2.34))
		})
		It("sets strings correctly", func() {
			target := &struct{ V string }{}
			Expect(setFromString(WrappedValue(target), "whoop")).To(Succeed())
			Expect(target.V).To(Equal("whoop"))
		})
		It("sets []string correctly", func() {
			target := &struct{ V []string }{}
			Expect(setFromString(WrappedValue(target), "wow,insane")).To(Succeed())
			Expect(target.V).To(Equal([]string{"wow", "insane"}))
		})
		It("sets map[string]string correctly", func() {
			target := &struct{ V map[string]string }{}
			Expect(setFromString(WrappedValue(target), "wow=insane")).To(Succeed())
			Expect(target.V).To(Equal(map[string]string{"wow": "insane"}))
		})
		It("sets TextUnmarshaler correctly", func() {
			target := &struct{ V testType }{}
			Expect(setFromString(WrappedValue(target), "mmh")).To(Succeed())
			Expect(target.V).To(Equal(testType{S: "mmh"}))
		})
	})

	Describe("readPFlags", func() {

	})

	Describe("readEnv", func() {

	})

	Describe("getEnvAsMap", func() {

	})

	Describe("readFiles", func() {

	})

	Describe("readFieldConfig", func() {

	})

	Describe("getFieldsConfigsFromValue", func() {

	})

	Describe("Collector", func() {
		Describe("Get", func() {

		})
	})
})

func WrappedValue(val interface{}) reflect.Value {
	return reflect.ValueOf(val).Elem().Field(0)
}

type testType struct {
	S string
}

func (t *testType) UnmarshalText(text []byte) error {
	t.S = string(text)

	return nil
}
