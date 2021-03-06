package alligotor

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("EnvSource", func() {
	//	Describe("readEnv", func() {
	//		var config EnvConfig
	//		BeforeEach(func() {
	//			config = EnvConfig{
	//				Prefix:    "",
	//				separator: "_",
	//				Disabled:  false,
	//			}
	//		})
	//		It("uses uppercase name as default env name", func() {
	//			err := readEnv(fields, config, map[string]string{"PORT": "3000"})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("uses configured name", func() {
	//			fields[0].Config.DefaultEnvName = "overwrite"
	//			err := readEnv(fields, config, map[string]string{"OVERWRITE": "3000"})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("uses prefix", func() {
	//			config.Prefix = "prefix"
	//			err := readEnv(fields, config, map[string]string{"PREFIX_PORT": "3000"})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("doesn't use prefix if name is configured", func() {
	//			config.Prefix = "prefix"
	//			fields[0].Config.DefaultEnvName = "overwrite"
	//			err := readEnv(fields, config, map[string]string{"OVERWRITE": "3000"})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("doesn't overwrite with empty value if not set", func() {
	//			target.V = 3000
	//			err := readEnv(fields, config, map[string]string{})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("overwrites with empty value if set to empty", func() {
	//			target.V = 3000
	//			err := readEnv(fields, config, map[string]string{"PORT": ""})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(0))
	//		})
	//		Context("nested", func() {
	//			It("uses separator", func() {
	//				err := readEnv(nestedFields, config, map[string]string{"SUB_PORT": "1234"})
	//				Expect(err).ShouldNot(HaveOccurred())
	//				Expect(nestedTarget.Sub.V).To(Equal(1234))
	//			})
	//			It("can be overridden", func() {
	//				nestedFields[0].Config.DefaultEnvName = "PORT"
	//				err := readEnv(nestedFields, config, map[string]string{"PORT": "1234"})
	//				Expect(err).ShouldNot(HaveOccurred())
	//				Expect(nestedTarget.Sub.V).To(Equal(1234))
	//			})
	//			It("uses distinct name instead of overridden/default if both are set", func() {
	//				nestedFields[0].Config.DefaultEnvName = "DEFAULT"
	//				err := readEnv(nestedFields, config, map[string]string{"DEFAULT": "1234", "SUB_PORT": "1235"})
	//				Expect(err).ShouldNot(HaveOccurred())
	//				Expect(nestedTarget.Sub.V).To(Equal(1235))
	//			})
	//			It("works if multiple fields are trying to get the same default flag", func() {
	//				nestedFields[0].Config.DefaultEnvName = "DEFAULT"
	//				nestedFields[1].Config.DefaultEnvName = "DEFAULT"
	//				err := readEnv(nestedFields, config, map[string]string{"DEFAULT": "1234", "SUB_PORT": "1235"})
	//				Expect(err).ShouldNot(HaveOccurred())
	//				Expect(nestedTarget.Sub.V).To(Equal(1235))
	//				Expect(nestedTarget.Sub.W).To(Equal(1234))
	//			})
	//		})
	//	})
	//Describe("getEnvAsMap", func() {
	//	It("gets environment variables in right format", func() {
	//		Expect(os.Setenv("TESTING_KEY", "TESTING_VAL")).To(Succeed())
	//		envMap := getEnvAsMap()
	//		testingVal, ok := envMap["TESTING_KEY"]
	//		Expect(ok).To(BeTrue())
	//		Expect(testingVal).To(Equal("TESTING_VAL"))
	//	})
	//	It("supports '=' and ',' in the value", func() {
	//		Expect(os.Setenv("TESTING_KEY", "lel=lol,arr=lul")).To(Succeed())
	//		envMap := getEnvAsMap()
	//		testingVal, ok := envMap["TESTING_KEY"]
	//		Expect(ok).To(BeTrue())
	//		Expect(testingVal).To(Equal("lel=lol,arr=lul"))
	//	})
	//})
})
