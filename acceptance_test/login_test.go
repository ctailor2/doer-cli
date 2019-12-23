package acceptance_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ctailor2/doer-cli/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("login", func() {
	var session *gexec.Session
	var server *ghttp.Server
	var cliPath string
	actionInput := "login"

	BeforeEach(func() {
		cliPath = buildCli()
		server = ghttp.NewServer()
		server.SetAllowUnhandledRequests(true)
		links := make(map[string]cmd.Link)
		links["login"] = cmd.Link{Href: strings.Join([]string{server.URL(), "loginHref"}, "/")}
		baseResources := cmd.ResourcesResponse{Links: links}
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, baseResources),
			),
		)
	})

	It("prompts the user for login email", func() {
		session = runCli(cliPath, actionInput, "--api", server.URL(), "--config", "test-config.yml")
		Expect(session).Should(gbytes.Say("Email"))
	})

	It("prompts the user for login password when user enters email", func() {
		input := gbytes.NewBuffer()
		_, err := input.Write([]byte("someEmail\n"))
		Expect(err).NotTo(HaveOccurred())
		session = runCliWithInput(cliPath, input, actionInput, "--api", server.URL(), "--config", "test-config.yml")
		Expect(session).Should(gbytes.Say("Password"))
	})

	It("attempts to login when user enters email and password", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/loginHref"),
				ghttp.VerifyJSON("{\"email\":\"someEmail\",\"password\":\"somePassword\"}"),
			),
		)
		input := gbytes.NewBuffer()
		_, err := input.Write([]byte("someEmail\n" + "somePassword\n"))
		Expect(err).NotTo(HaveOccurred())
		session = runCliWithInput(cliPath, input, actionInput, "--api", server.URL(), "--config", "test-config.yml")
		Expect(server.ReceivedRequests()).Should(HaveLen(2))
	})

	It("writes the session token and root resources href to config file when login successful", func() {
		links := make(map[string]cmd.Link)
		links["root"] = cmd.Link{Href: "rootResourcesHref"}
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/loginHref"),
				ghttp.RespondWithJSONEncoded(200, cmd.SessionResponse{
					Session: cmd.Session{
						Token: "someToken",
					},
					Links: links,
				}),
			),
		)
		input := gbytes.NewBuffer()
		_, err := input.Write([]byte("someEmail\n" + "somePassword\n"))
		Expect(err).NotTo(HaveOccurred())
		session = runCliWithInput(cliPath, input, actionInput, "--api", server.URL(), "--config", "test-config.yml")
		Expect(server.ReceivedRequests()).Should(HaveLen(2))
		contents, _ := ioutil.ReadFile("test-config.yml")
		contentString := string(contents)
		Expect(contentString).To(ContainSubstring("someToken"))
		Expect(contentString).To(ContainSubstring("rootResourcesHref"))
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
		os.Remove("./test-config.yml")
		server.Close()
	})
})
