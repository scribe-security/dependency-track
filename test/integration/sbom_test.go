package integration

import (
	"deptrack/client"
	"deptrack/core"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"gotest.tools/assert"
)

func TestCycloneDxRead(t *testing.T) {
	tests := []struct {
		fixture    string
		components int
	}{
		{
			fixture:    "test-fixtures/sbom/python.sbom.json",
			components: 25692,
		},
	}

	cyclonedx_manager, err := core.NewCycloneDxManager(core.JSON_FORMAT)
	if err != nil {
		t.Fatalf("Cyclonedx manager failed %+v", err)
	}
	for _, test := range tests {
		t.Run(test.fixture, func(t *testing.T) {
			var bom cdx.BOM
			err := cyclonedx_manager.ReadFromFile(test.fixture, &bom)
			if err != nil {
				t.Fatalf("Cyclonedx manager failed %+v", err)
			}
			assert.Equal(t, test.components, len(*bom.Components))
		})
	}
}

func TestCycloneDxPost(t *testing.T) {
	tests := []struct {
		fixture    string
		components int
	}{
		{
			fixture:    "test-fixtures/sbom/python.sbom.json",
			components: 25692,
		},
	}

	c := GetLocalDepClient(t)
	cyclonedx_manager, err := core.NewCycloneDxManager(core.JSON_FORMAT)
	if err != nil {
		t.Fatalf("Cyclonedx manager failed %+v", err)
	}
	for _, test := range tests {
		t.Run(test.fixture, func(t *testing.T) {
			var bom cdx.BOM
			err = cyclonedx_manager.ReadFromFile(test.fixture, &bom)
			if err != nil {
				t.Fatalf("Read from file: %+v", err)
			}

			name, err := cyclonedx_manager.GetName(&bom)
			if err != nil {
				t.Fatalf("Get sbom name: %+v", err)
			}

			params := client.DepTrackSbomPost{
				AutoCreate:  "true",
				ProjectName: name,
			}

			var sbom_response client.DepTrackSbomPostResponse
			err = c.PostSbom("bom", &params, &bom, &sbom_response)
			if err != nil {
				t.Fatalf("Get response: %+v", err)
			}

			if sbom_response.Token == "" {
				t.Fatalf("Token not received")
			}
			t.Log(sbom_response)
		})
	}
}
