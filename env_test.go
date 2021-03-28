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
				name: name,
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
			field.configs = map[string]string{envKey: "overwrite"}

			val := readEnv(field, "", map[string]string{"OVERWRITE": "3000"}, separator)
			Expect(val).To(Equal([]byte("3000")))
		})
		Context("prefix", func() {
			It("uses prefix", func() {
				val := readEnv(field, "prefix", map[string]string{"PREFIX" + separator + strings.ToUpper(name): "3000"}, separator)
				Expect(val).To(Equal([]byte("3000")))
			})
			It("allows overwriting names", func() {
				field.configs = map[string]string{envKey: "overwrite"}
				val := readEnv(field, "prefix", map[string]string{"PREFIX" + separator + "OVERWRITE": "3000"}, separator)
				Expect(val).To(Equal([]byte("3000")))
			})
			It("uses overwritten name even if normal one is set", func() {
				field.configs = map[string]string{envKey: "overwrite"}
				val := readEnv(
					field, "prefix",
					map[string]string{
						"PREFIX" + separator + strings.ToUpper(name): "2000",
						"PREFIX" + separator + "OVERWRITE":           "3000"},
					separator,
				)
				Expect(val).To(Equal([]byte("3000")))
			})
		})
		Context("nested", func() {
			var base = "base"

			BeforeEach(func() {
				field.base = []Field{{name: base}}
			})
			It("uses separator", func() {
				val := readEnv(field, "", map[string]string{strings.ToUpper(base + separator + name): "1234"}, separator)
				Expect(val).To(Equal([]byte("1234")))
			})
			It("can be overwritten", func() {
				field.configs = map[string]string{envKey: "overwrite"}
				val := readEnv(field, "", map[string]string{strings.ToUpper(base + separator + "overwrite"): "1234"}, separator)
				Expect(val).To(Equal([]byte("1234")))
			})
			It("uses overwritten name even if normal one is set", func() {
				field.configs = map[string]string{envKey: "overwrite"}
				val := readEnv(field, "", map[string]string{
					strings.ToUpper(base + separator + name):        "1234",
					strings.ToUpper(base + separator + "overwrite"): "1235",
				}, separator)
				Expect(val).To(Equal([]byte("1235")))
			})
			It("supports overwriting base", func() {
				field.base = []Field{{name: base, configs: map[string]string{envKey: "overwrittenbase"}}}
				val := readEnv(field, "", map[string]string{strings.ToUpper("overwrittenbase" + separator + name): "1234"}, separator)
				Expect(val).To(Equal([]byte("1234")))
			})
		})
	})
})
