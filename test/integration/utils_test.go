package integration

import (
	"deptrack/client"
	"deptrack/core"
	"os"
	"testing"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"gotest.tools/assert"
)

func GetLocalDepClient(t *testing.T) *client.DepTrackClient {
	api_key, ok := os.LookupEnv("API_KEY")
	if !ok {
		t.Fatalf("No api key")
	}
	c, err := client.NewDepTrackClient(api_key)
	assert.NilError(t, err, "Failed to create client")
	return c
}

func GetCycloneDxManager(t *testing.T) *core.CycloneDxManager {
	cyclonedx_manager, err := core.NewCycloneDxManager(core.JSON_FORMAT)
	assert.NilError(t, err, "Cyclonedx manager create")
	return cyclonedx_manager
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

func BasePostLogic(t *testing.T, path string) (*client.DepTrackClient, *core.CycloneDxManager, *cdx.BOM, *client.DepTrackSbomPostResponse) {
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
