package alligotor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("flags", func() {
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
					Expect(f).To(Equal(flag{ShortName: "a", LongName: "awd"}))
				}
			})
			It("should return valid flag when only short is set", func() {
				f, err := readFlagConfig("a")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(f).To(Equal(flag{ShortName: "a", LongName: ""}))
			})
			It("should return valid flag when only long is set", func() {
				f, err := readFlagConfig("awd")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(f).To(Equal(flag{ShortName: "", LongName: "awd"}))
			})
		})
	})
	Describe("FlagsSource", func() {
		var (
			s         *FlagsSource
			field     *Field
			fields    []*Field
			separator = "-"
			name      = "name"
			flagName  = "--" + name
		)
		BeforeEach(func() {
			s = &FlagsSource{
				Separator:       separator,
				fieldToFlagInfo: map[*Field]*flagInfo{},
			}
			field = &Field{Name: name}
			fields = []*Field{field}
		})
		Describe("initFlagMap", func() {
			It("contains flag for field", func() {
				Expect(s.initFlagMap(fields, []string{flagName, "3000"})).To(Succeed())

				flagInfo, ok := s.fieldToFlagInfo[field]
				Expect(ok).To(BeTrue())
				Expect(*flagInfo.valueStr).To(Equal("3000"))
			})
			It("supports overwriting longname", func() {
				field.Configs = map[string]string{flagKey: "overwrite"}
				Expect(s.initFlagMap(fields, []string{flagName, "3000", "--overwrite", "4000"})).To(Succeed())

				flagInfo, ok := s.fieldToFlagInfo[field]
				Expect(ok).To(BeTrue())
				Expect(*flagInfo.valueStr).To(Equal("4000"))
			})
		})
		Describe("Read", func() {
			It("returns nil if not set", func() {
				Expect(s.initFlagMap(fields, []string{})).To(Succeed())
				val, err := s.Read(field)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(BeNil())
			})
			It("returns empty string if set to empty string", func() {
				Expect(s.initFlagMap(fields, []string{flagName, ""})).To(Succeed())
				val, err := s.Read(field)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal([]byte("")))
			})
			It("uses name as normal flag name", func() {
				Expect(s.initFlagMap(fields, []string{flagName, "3000"})).To(Succeed())
				val, err := s.Read(field)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal([]byte("3000")))
			})
			It("uses configured long name", func() {
				field.Configs = map[string]string{flagKey: "overwrite"}
				Expect(s.initFlagMap(fields, []string{"--overwrite", "3000"})).To(Succeed())
				val, err := s.Read(field)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal([]byte("3000")))
			})
			It("uses configured short name", func() {
				field.Configs = map[string]string{flagKey: "o"}
				Expect(s.initFlagMap(fields, []string{"-o", "3000"})).To(Succeed())
				val, err := s.Read(field)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal([]byte("3000")))
			})
			Context("nested", func() {
				var base = "base"
				BeforeEach(func() {
					field.Base = []string{base}
					flagName = "--" + base + separator + name
				})
				It("uses separator", func() {
					Expect(s.initFlagMap(fields, []string{flagName, "3000"})).To(Succeed())
					val, err := s.Read(field)
					Expect(err).ToNot(HaveOccurred())
					Expect(val).To(Equal([]byte("3000")))
				})
				It("can use defaults", func() {
					field.Configs = map[string]string{flagKey: "overwrite"}
					Expect(s.initFlagMap(fields, []string{"--" + base + separator + "overwrite", "3000"})).To(Succeed())
					val, err := s.Read(field)
					Expect(err).ToNot(HaveOccurred())
					Expect(val).To(Equal([]byte("3000")))
				})
				It("uses overridden name even if normal one is set", func() {
					field.Configs = map[string]string{flagKey: "overwrite"}
					Expect(s.initFlagMap(fields, []string{
						"--" + base + separator + "overwrite", "1234",
						flagName, "1235"},
					)).To(Succeed())
					val, err := s.Read(field)
					Expect(err).ToNot(HaveOccurred())
					Expect(val).To(Equal([]byte("1234")))
				})
			})
		})
	})

})
