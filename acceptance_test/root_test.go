package acceptance_test

import (
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

var _ = Describe("root", func() {
	var session *gexec.Session
	var server *ghttp.Server
	var cliPath string

	BeforeEach(func() {
		cliPath = buildCli()
		server = ghttp.NewServer()
		links := make(map[string]cmd.Link)
		links["self"] = cmd.Link{Href: strings.Join([]string{server.URL(), "selfHref"}, "/")}
		links["dummyOption"] = cmd.Link{Href: strings.Join([]string{server.URL(), "dummyOptionHref"}, "/")}
		links["login"] = cmd.Link{Href: strings.Join([]string{server.URL(), "loginHref"}, "/")}
		links["signup"] = cmd.Link{Href: strings.Join([]string{server.URL(), "signupHref"}, "/")}
		baseResources := cmd.ResourcesResponse{Links: links}
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, baseResources),
			),
		)
	})

	When("session token is not set", func() {
		It("presents a selection from resources other than self", func() {
			session := runCli(cliPath, "--api", server.URL(), "--config", "test-config.yml")
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
			Expect(session).Should(gbytes.Say("Choose action"))
			Expect(session).Should(gbytes.Say("dummyOption"))
			Expect(session).Should(gbytes.Say("login"))
			Expect(session).Should(gbytes.Say("signup"))
			Expect(session).ShouldNot(gbytes.Say("self"))
		})
	})

	When("session token is set", func() {
		BeforeEach(func() {
			links := make(map[string]cmd.Link)
			links["root"] = cmd.Link{Href: server.URL() + "/rootResourcesHref"}
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
			session = runCliWithInput(cliPath, input, "login", "--api", server.URL(), "--config", "test-config.yml")
		})

		It("fetches the root resources using the session token", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/rootResourcesHref"),
					ghttp.VerifyHeaderKV("Session-Token", "someToken"),
				),
			)
			runCli(cliPath, "--api", server.URL(), "--config", "test-config.yml")
			Expect(server.ReceivedRequests()).Should(HaveLen(3))
		})

		It("presents a selection from resources other than self", func() {
			links := make(map[string]cmd.Link)
			links["self"] = cmd.Link{Href: "selfHref"}
			links["rootResource1"] = cmd.Link{Href: "rootResource1Href"}
			links["rootResource2"] = cmd.Link{Href: "rootResource2Href"}
			rootResources := cmd.ResourcesResponse{Links: links}
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/rootResourcesHref"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, rootResources),
				),
			)
			session = runCli(cliPath, "--api", server.URL(), "--config", "test-config.yml")
			Expect(session).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say("Choose action"))
			Expect(session).Should(gbytes.Say("rootResource1"))
			Expect(session).Should(gbytes.Say("rootResource2"))
			Expect(session).ShouldNot(gbytes.Say("self"))
		})
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
		os.Remove("./test-config.yml")
		server.Close()
	})
})
