package alligotor

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("env", func() {
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
	Describe("readEnv", func() {
		var (
			field     *Field
			separator = "_"
			name      = "someInt"
		)
		BeforeEach(func() {
			field = &Field{
				Name: name,
			}
		})
		It("returns nil if not set", func() {
			val := readEnv(field, "", nil, separator)
			Expect(val).To(BeNil())
		})
		It("returns empty string if set to empty string", func() {
			val := readEnv(field, "", map[string]string{strings.ToUpper(name): ""}, separator)
			Expect(val).To(Equal([]byte("")))
		})
		It("uses uppercase name as default env name", func() {
			val := readEnv(field, "", map[string]string{strings.ToUpper(name): "3000"}, separator)
			Expect(val).To(Equal([]byte("3000")))
		})
		It("uses configured name", func() {
			field.Configs = map[string]string{envKey: "overwrite"}

			val := readEnv(field, "", map[string]string{"OVERWRITE": "3000"}, separator)
			Expect(val).To(Equal([]byte("3000")))
		})
		It("uses prefix", func() {
			val := readEnv(field, "prefix", map[string]string{"PREFIX" + separator + strings.ToUpper(name): "3000"}, separator)
			Expect(val).To(Equal([]byte("3000")))
		})
		It("doesn't use prefix if name is configured", func() {
			field.Configs = map[string]string{envKey: "overwrite"}
			val := readEnv(field, "prefix", map[string]string{"OVERWRITE": "3000"}, separator)
			Expect(val).To(Equal([]byte("3000")))
		})
		It("uses disting name over long name", func() {
			field.Configs = map[string]string{envKey: "overwrite"}
			val := readEnv(
				field, "prefix",
				map[string]string{
					"PREFIX" + separator + strings.ToUpper(name): "2000",
					"OVERWRITE": "3000"},
				separator,
			)
			Expect(val).To(Equal([]byte("2000")))
		})
		Context("nested", func() {
			var (
				base    = "base"
				envName = strings.ToUpper(base + separator + name)
			)
			BeforeEach(func() {
				field.Base = []string{base}
			})
			It("uses separator", func() {
				val := readEnv(field, "", map[string]string{envName: "1234"}, separator)
				Expect(val).To(Equal([]byte("1234")))
			})
			It("can be overridden", func() {
				field.Configs = map[string]string{envKey: "overwrite"}
				val := readEnv(field, "", map[string]string{"OVERWRITE": "1234"}, separator)
				Expect(val).To(Equal([]byte("1234")))
			})
			It("uses distinct name instead of overridden/default if both are set", func() {
				field.Configs = map[string]string{envKey: "overwrite"}
				val := readEnv(field, "", map[string]string{"OVERWRITE": "1234", envName: "1235"}, separator)
				Expect(val).To(Equal([]byte("1235")))
			})
		})
	})
})
