package acceptance_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("login", func() {
	var session *gexec.Session
	var server *ghttp.Server

	BeforeEach(func() {
		cliPath := buildCli()
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.RespondWith(200, ""),
		)
		session = runCli(cliPath, server.URL())
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
		server.Close()
	})

	It("exits with status code 0", func() {
		Eventually(session).Should(gexec.Exit(0))
	})

	It("prompts the user for login email", func() {
		Eventually(session).Should(gbytes.Say("Email"))
	})

	It("prompts the user for login password when user enters email", func() {
		var buffer = gbytes.NewBuffer()
		_, err := buffer.Write([]byte("someEmail\n"))
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Password"))
	})

	It("attempts to login when user enters email and password", func() {
		server.AppendHandlers(
			ghttp.VerifyRequest("POST", "/v1/login"),
			ghttp.VerifyJSON("{\"email\":\"someEmail\",\"password\":\"somePassword\"}"),
		)
		var buffer = gbytes.NewBuffer()
		_, emailErr := buffer.Write([]byte("someEmail\n"))
		Expect(emailErr).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Password"))
		_, passwordErr := buffer.Write([]byte("somePassword\n"))
		Expect(passwordErr).NotTo(HaveOccurred())
		Eventually(server.ReceivedRequests()).Should(HaveLen(1))
	})
})

func buildCli() string {
	cliPath, err := gexec.Build("github.com/ctailor2/doer-cli")
	Expect(err).NotTo(HaveOccurred())

	return cliPath
}

func runCli(path string, serverUrl string) *gexec.Session {
	cmd := exec.Command(path, "login", "--target", serverUrl)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
