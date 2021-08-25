package integration

import (
	"deptrack/client"
	"os"
	"testing"
	"time"

	cdx_manager "github.com/scribe-security/scribe/pkg/cyclonedx"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"gotest.tools/assert"
)

const ApiServerPath = "http://localhost:8081/api/v1"

func GetLocalDepClient(t *testing.T) *client.DepTrackClient {
	api_key, ok := os.LookupEnv("API_KEY")
	if !ok {
		t.Fatalf("No api key")
	}
	c, err := client.NewDepTrackClient(api_key, ApiServerPath)
	assert.NilError(t, err, "Failed to create client")
	return c
}

func GetCycloneDxManager(t *testing.T) *cdx_manager.CycloneDxManager {
	cyclonedx_manager, err := cdx_manager.NewCycloneDxManager(cdx_manager.JSON_FORMAT)
	assert.NilError(t, err, "Cyclonedx manager create")
	return cyclonedx_manager
}

func ReadSbom(t *testing.T, fixture string, m *cdx_manager.CycloneDxManager) (*cdx.BOM, string) {
	var bom cdx.BOM
	err := m.ReadFromFile(fixture, &bom)
	assert.NilError(t, err, "Read from sbom")

	name, err := m.GetName(&bom)
	assert.NilError(t, err, "Get sbom name")

	return &bom, name
}

func PostSbom(t *testing.T, name string, c *client.DepTrackClient, bom *cdx.BOM) *client.DepTrackSbomPostResponse {
	params := client.DepTrackSbomPost{
		AutoCreate:  "true",
		ProjectName: name,
	}
	var sbom_response client.DepTrackSbomPostResponse
	err := c.PostSbom("bom", &params, bom, &sbom_response)
	assert.NilError(t, err, "Failed to post sbom")
	assert.Assert(t, sbom_response.Token != "", "Empty token")
	return &sbom_response
}

func BasePostLogic(t *testing.T, path string) (*client.DepTrackClient, *cdx_manager.CycloneDxManager, *cdx.BOM, *client.DepTrackSbomPostResponse) {
	c := GetLocalDepClient(t)
	cyclonedx_manager := GetCycloneDxManager(t)
	bom, name := ReadSbom(t, path, cyclonedx_manager)
	sbom_response := PostSbom(t, name, c, bom)

	start := time.Now()
	_, err := c.WaitforSbomFinishUpload(sbom_response.Token)
	assert.NilError(t, err, "Wait for sbom upload")

	time := time.Since(start)
	t.Log("Wait time: ", time)
	return c, cyclonedx_manager, bom, sbom_response
}
