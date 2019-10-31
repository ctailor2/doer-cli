package acceptance_test

import (
	"os/exec"
	"github.com/onsi/gomega/gexec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAcceptanceTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AcceptanceTest Suite")
}

func buildCli() string {
	cliPath, err := gexec.Build("github.com/ctailor2/doer-cli")
	Expect(err).NotTo(HaveOccurred())

	return cliPath
}

func runCli(path string, args ...string) *gexec.Session {
	cmd := exec.Command(path, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
