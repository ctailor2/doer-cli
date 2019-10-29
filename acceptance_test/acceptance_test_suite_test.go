package acceptance_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAcceptanceTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AcceptanceTest Suite")
}
