package acceptance_test

import (
	"os"
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
			ghttp.RespondWith(200, ""),
		)
		session = runCli(cliPath, "--api", server.URL(), "--config", "test-config.yml")
		Eventually(session).Should(gexec.Exit(0))
		session = runCli(cliPath, "login", "--config", "test-config.yml")
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
		os.Remove("./test-config.yml")
		server.Close()
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
		_, err := buffer.Write([]byte("someEmail\n"))
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Password"))
		_, err = buffer.Write([]byte("somePassword\n"))
		Expect(err).NotTo(HaveOccurred())
		Eventually(server.ReceivedRequests()).Should(HaveLen(2))
	})
})
