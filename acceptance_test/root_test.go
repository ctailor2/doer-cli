package acceptance_test

import (
	"net/http"
	"os"

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
		server.SetAllowUnhandledRequests(true)
		links := make(map[string]cmd.Link)
		links["self"] = cmd.Link{Href: "selfHref"}
		links["baseResource1"] = cmd.Link{Href: "baseResource1Href"}
		links["baseResource2"] = cmd.Link{Href: "baseResource2Href"}
		baseResources := cmd.ResourcesResponse{Links: links}
		server.AppendHandlers(
			ghttp.RespondWithJSONEncoded(http.StatusOK, baseResources),
		)
		session = runCli(cliPath, "--api", server.URL(), "--config", "test-config.yml")
	})

	When("session token is not set", func() {
		It("fetches the base resources", func() {
			Eventually(session).Should(gexec.Exit(0))
			server.AppendHandlers(
				ghttp.VerifyRequest("GET", "/v1/"),
			)
			Eventually(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("presents a selection from resources other than self", func() {
			Eventually(session).Should(gbytes.Say("Choose:"))
			Eventually(session).Should(gbytes.Say("baseResource1"))
			Eventually(session).Should(gbytes.Say("baseResource2"))
			Eventually(session).ShouldNot(gbytes.Say("self"))
		})
	})

	When("session token is set", func() {
		BeforeEach(func() {
			Eventually(session).Should(gexec.Exit(0))
			links := make(map[string]cmd.Link)
			links["root"] = cmd.Link{Href: server.URL() + "/rootResourcesHref"}
			server.AppendHandlers(
				ghttp.RespondWithJSONEncoded(200, cmd.SessionResponse{
					Session: cmd.Session{
						Token: "someToken",
					},
					Links: links,
				}),
			)
			session = runCli(cliPath, "login", "--config", "test-config.yml")
			var buffer = gbytes.NewBuffer()
			_, err := buffer.Write([]byte("someEmail\n"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("Password"))
			_, err = buffer.Write([]byte("somePassword\n"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("fetches the root resources using the session token", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/rootResourcesHref"),
					ghttp.VerifyHeaderKV("Session-Token", "someToken"),
				),
			)
			session = runCli(cliPath, "--api", server.URL(), "--config", "test-config.yml")
			Eventually(session).Should(gexec.Exit(0))
			Eventually(server.ReceivedRequests()).Should(HaveLen(3))
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
			Eventually(session).Should(gexec.Exit(0))
			Eventually(session).Should(gbytes.Say("Choose:"))
			Eventually(session).Should(gbytes.Say("rootResource1"))
			Eventually(session).Should(gbytes.Say("rootResource2"))
			Eventually(session).ShouldNot(gbytes.Say("self"))
		})
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
		os.Remove("./test-config.yml")
		server.Close()
	})
})
