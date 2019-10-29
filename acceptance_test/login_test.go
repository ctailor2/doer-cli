package acceptance_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("something", func() {
	var session *gexec.Session

	BeforeEach(func() {
		cliPath := buildCli()
		session = runCli(cliPath)
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("exits with status code 0", func() {
		Eventually(session).Should(gexec.Exit(0))
	})

	It("prints 'Hello World' to stdout", func() {
		Eventually(session).Should(gbytes.Say("Hello World"))
	})
})

func buildCli() string {
	cliPath, err := gexec.Build("github.com/ctailor2/doer-cli")
	Expect(err).NotTo(HaveOccurred())

	return cliPath
}

func runCli(path string) *gexec.Session {
	cmd := exec.Command(path)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
