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

var _ = Describe("signup", func() {
	var session *gexec.Session
	var server *ghttp.Server
	var cliPath string
	actionInput := "signup"

	BeforeEach(func() {
		cliPath = buildCli()
		server = ghttp.NewServer()
		server.SetAllowUnhandledRequests(true)
		links := make(map[string]cmd.Link)
		links["signup"] = cmd.Link{Href: strings.Join([]string{server.URL(), "signupHref"}, "/")}
		baseResources := cmd.ResourcesResponse{Links: links}
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, baseResources),
			),
		)
	})

	It("prompts the user for signup email", func() {
		session = runCli(cliPath, actionInput, "--api", server.URL(), "--config", "test-config.yml")
		Expect(session).Should(gbytes.Say("Email"))
	})

	It("prompts the user for signup password when user enters email", func() {
		input := gbytes.NewBuffer()
		_, err := input.Write([]byte("someEmail\n"))
		Expect(err).NotTo(HaveOccurred())
		session = runCliWithInput(cliPath, input, actionInput, "--api", server.URL(), "--config", "test-config.yml")
		Expect(session).Should(gbytes.Say("Password"))
	})

	It("prompts the user for signup password confirmation when user enters password", func() {
		input := gbytes.NewBuffer()
		_, err := input.Write([]byte("someEmail\n" + "somePassword\n"))
		Expect(err).NotTo(HaveOccurred())
		session = runCliWithInput(cliPath, input, actionInput, "--api", server.URL(), "--config", "test-config.yml")
		Expect(session).Should(gbytes.Say("Password Confirmation"))
	})

	When("password confirmation matches password", func() {
		var input *gbytes.Buffer

		BeforeEach(func() {
			input = gbytes.NewBuffer()
			_, err := input.Write([]byte("someEmail\n" + "somePassword\n" + "somePassword\n"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("attempts to signup", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/signupHref"),
					ghttp.VerifyJSON("{\"email\":\"someEmail\",\"password\":\"somePassword\"}"),
				),
			)
			session = runCliWithInput(cliPath, input, actionInput, "--api", server.URL(), "--config", "test-config.yml")
			Expect(server.ReceivedRequests()).Should(HaveLen(2))
		})

		It("writes the session token and root resources href to config file when signup successful", func() {
			links := make(map[string]cmd.Link)
			links["root"] = cmd.Link{Href: "rootResourcesHref"}
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/signupHref"),
					ghttp.RespondWithJSONEncoded(200, cmd.SessionResponse{
						Session: cmd.Session{
							Token: "someToken",
						},
						Links: links,
					}),
				),
			)
			session = runCliWithInput(cliPath, input, actionInput, "--api", server.URL(), "--config", "test-config.yml")
			Expect(server.ReceivedRequests()).Should(HaveLen(2))
			contents, _ := ioutil.ReadFile("test-config.yml")
			contentString := string(contents)
			Expect(contentString).To(ContainSubstring("someToken"))
			Expect(contentString).To(ContainSubstring("rootResourcesHref"))
		})
	})

	When("password confirmation does not match password", func() {
		var input *gbytes.Buffer

		BeforeEach(func() {
			input = gbytes.NewBuffer()
			_, err := input.Write([]byte("someEmail\n" + "somePassword\n" + "someOtherPassword\n"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("displays an error", func() {
			session = runCliWithInput(cliPath, input, actionInput, "--api", server.URL(), "--config", "test-config.yml")
			Expect(session).To(gbytes.Say("Password confirmation and password do not match."))
		})
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
		os.Remove("./test-config.yml")
		server.Close()
	})
})
