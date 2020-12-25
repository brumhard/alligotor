package alligotor_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAlligotor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alligotor Suite")
}
