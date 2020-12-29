package alligotor

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ciMap", func() {
	var (
		ciMap     *ciMap
		jsonBytes = []byte(`{
	"test": {
		"innertest": "arrrr",
		"innertest2": "pirate"
	},
	"test2": "idk"
}`)
	)
	BeforeEach(func() {
		ciMap = newCiMap()
		Expect(json.Unmarshal(jsonBytes, ciMap)).To(Succeed())
	})
	Describe("Get", func() {
		Context("nested", func() {
			It("works", func() {
				val, ok := ciMap.Get("test" + defaultSeparator + "innertest")
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal("arrrr"))
			})
		})
		Context("case insensitive", func() {
			It("works", func() {
				val, ok := ciMap.Get("test" + defaultSeparator + "INNERTEST2")
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal("pirate"))
			})
		})
		Context("in root", func() {
			It("works", func() {
				val, ok := ciMap.Get("TEST2")
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal("idk"))
			})
		})
		Context("key does not exist", func() {
			It("should return ok=false", func() {
				_, ok := ciMap.Get("not-existing")
				Expect(ok).To(BeFalse())
			})
		})
	})
})
