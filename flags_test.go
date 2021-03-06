package alligotor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FlagsSource", func() {
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
					f, err := readFlagConfig(configStr)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(f).To(Equal(flag{ShortName: "a", DefaultName: "awd"}))
				}
			})
			It("should return valid flag when only short is set", func() {
				f, err := readFlagConfig("a")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(f).To(Equal(flag{ShortName: "a", DefaultName: ""}))
			})
			It("should return valid flag when only long is set", func() {
				f, err := readFlagConfig("awd")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(f).To(Equal(flag{ShortName: "", DefaultName: "awd"}))
			})
		})
	})
	//	Describe("readPFlags", func() {
	//		config := FlagsConfig{
	//			separator: "-",
	//			Disabled:  false,
	//		}
	//
	//		It("uses name as default flag name", func() {
	//			err := readPFlags(fields, config, []string{"--port", "3000"})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("uses configured long name", func() {
	//			fields[0].Config.Flag.DefaultName = "overwrite"
	//			err := readPFlags(fields, config, []string{"--overwrite", "3000"})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("uses configured short name", func() {
	//			fields[0].Config.Flag.ShortName = "o"
	//			err := readPFlags(fields, config, []string{"-o", "3000"})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("doesn't overwrite with empty value if not set", func() {
	//			target.V = 3000
	//			err := readPFlags(fields, config, []string{})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(3000))
	//		})
	//		It("overwrites with empty value if set to empty", func() {
	//			target.V = 3000
	//			err := readPFlags(fields, config, []string{"--port", ""})
	//			Expect(err).ShouldNot(HaveOccurred())
	//			Expect(target.V).To(Equal(0))
	//		})
	//		Context("nested", func() {
	//			It("uses separator", func() {
	//				err := readPFlags(nestedFields, config, []string{"--sub-port", "1234"})
	//				Expect(err).ShouldNot(HaveOccurred())
	//				Expect(nestedTarget.Sub.V).To(Equal(1234))
	//			})
	//			It("can use defaults", func() {
	//				nestedFields[0].Config.Flag.DefaultName = "default"
	//				err := readPFlags(nestedFields, config, []string{"--default", "1234"})
	//				Expect(err).ShouldNot(HaveOccurred())
	//				Expect(nestedTarget.Sub.V).To(Equal(1234))
	//			})
	//			It("uses distinct name instead of overridden/default if both are set", func() {
	//				nestedFields[0].Config.Flag.DefaultName = "default"
	//				err := readPFlags(nestedFields, config, []string{"--default", "1234", "--sub-port", "1235"})
	//				Expect(err).ShouldNot(HaveOccurred())
	//				Expect(nestedTarget.Sub.V).To(Equal(1235))
	//			})
	//			It("works if multiple fields are trying to get the same default flag", func() {
	//				nestedFields[0].Config.Flag.DefaultName = "default"
	//				nestedFields[1].Config.Flag.DefaultName = "default"
	//				err := readPFlags(nestedFields, config, []string{"--default", "1234", "--sub-port", "1235"})
	//				Expect(err).ShouldNot(HaveOccurred())
	//				Expect(nestedTarget.Sub.V).To(Equal(1235))
	//				Expect(nestedTarget.Sub.W).To(Equal(1234))
	//			})
	//		})
	//	})
})
