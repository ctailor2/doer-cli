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

	BeforeEach(func() {
		cliPath := buildCli()
		server = ghttp.NewServer()
		links := make(map[string]cmd.Link)
		links["self"] = cmd.Link{Href: "selfHref"}
		links["baseResource1"] = cmd.Link{Href: "baseResource1Href"}
		links["baseResource2"] = cmd.Link{Href: "baseResource2Href"}
		baseResources := cmd.BaseResourcesResponse{Links: links}
		server.AppendHandlers(
			ghttp.RespondWithJSONEncoded(http.StatusOK, baseResources),
		)
		session = runCli(cliPath, "--api", server.URL(), "--config", "test-config.yml")
	})

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

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
		os.Remove("./test-config.yml")
		server.Close()
	})
})
