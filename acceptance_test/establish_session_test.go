package acceptance_test

import (
	"io/ioutil"
	"github.com/ctailor2/doer-cli/cmd"
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
		links := make(map[string]cmd.Link)
		links["root"] = cmd.Link{Href: "rootResourcesHref"}
		server.AppendHandlers(
			ghttp.RespondWith(200, ""),
			ghttp.RespondWithJSONEncoded(200, cmd.SessionResponse{
				Session: cmd.Session{
					Token: "someToken",
				},
				Links: links,
			}),
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

	It("writes the session token and root resources href to config file when login successful", func() {
		var buffer = gbytes.NewBuffer()
		_, err := buffer.Write([]byte("someEmail\n"))
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Password"))
		_, err = buffer.Write([]byte("somePassword\n"))
		Expect(err).NotTo(HaveOccurred())
		Eventually(server.ReceivedRequests()).Should(HaveLen(2))
		contents, _ := ioutil.ReadFile("test-config.yml")
		contentString := string(contents)
		Expect(contentString).To(ContainSubstring("someToken"))
		Expect(contentString).To(ContainSubstring("rootResourcesHref"))
	})
})
