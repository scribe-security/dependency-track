package integration

import (
	"deptrack/client"
	"deptrack/core"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"gotest.tools/assert"
)

func UNUSED(x ...interface{}) {}

func ReadSbom(t *testing.T, fixture string, m *core.CycloneDxManager) (*cdx.BOM, string) {
	var bom cdx.BOM
	err := m.ReadFromFile(fixture, &bom)
	assert.NilError(t, err, "Read from sbom")

	name, err := m.GetName(&bom)
	assert.NilError(t, err, "Get sbom name")

	return &bom, name
}

func TestCycloneDxRead(t *testing.T) {
	tests := []struct {
		fixture    string
		name       string
		components int
	}{
		{
			fixture:    "test-fixtures/sbom/python.sbom.json",
			name:       "python:latest",
			components: 25692,
		},
	}

	cyclonedx_manager := GetCycloneDxManager(t)
	for _, test := range tests {
		t.Run(test.fixture, func(t *testing.T) {
			var bom cdx.BOM
			err := cyclonedx_manager.ReadFromFile(test.fixture, &bom)
			assert.NilError(t, err, "Cyclonedx manager read from file")

			assert.Equal(t, test.components, len(*bom.Components))

			name, err := cyclonedx_manager.GetName(&bom)
			assert.NilError(t, err, "Get sbom name")
			assert.Equal(t, name, test.name)
		})
	}
}

type test_purl_fixtures struct {
	purl_name           string
	purl_name_With_cves string
	component           int
	current             string
	latest              string
	vul_num             int
}

type test_sbom_and_check_fixtures struct {
	fixture          string
	components_group string
	components       int
	purl_fixtures    []test_purl_fixtures
}

func TestSbomByPurl(t *testing.T) {

	tests := test_sbom_and_check_fixtures{
		fixture:          "test-fixtures/sbom/python.sbom.json",
		components_group: "python",
		components:       6,
		purl_fixtures: []test_purl_fixtures{
			{
				purl_name:           "pkg:pypi/argparse@1.2.1",
				purl_name_With_cves: "pkg:deb/debian/git@1%3A2.20.1-2%20deb10u3?arch=amd64",
				current:             "1.2.1",
				latest:              "1.4.0",
				vul_num:             21,
				component:           1,
			},
		},
	}

	c, _, _, _ := BasePostLogic(t, tests.fixture)
	components, err := c.GetComponentsIdentity(client.GetComponentsIdentityParams{Group: tests.components_group, PaginationParams: client.DefaultPagination})
	assert.NilError(t, err, "Get components identity by group")
	assert.Equal(t, tests.components, len(components))

	for _, test := range tests.purl_fixtures {
		component, err := c.GetComponentsIdentity(client.GetComponentsIdentityParams{Purl: test.purl_name})
		assert.NilError(t, err, "Get repository identity by purl")
		assert.Equal(t, len(component), test.component)

		latest, err := c.GetRepositoryLatest(test.purl_name)
		assert.NilError(t, err, "Get repository latest")
		assert.Equal(t, latest.LatestVersion, test.latest)

		latestVersion, currentVersion, isVersionEquel, err := c.GetLatestVersion(test.purl_name)
		assert.NilError(t, err, "Get latest version")
		assert.Equal(t, currentVersion.Version, test.current)
		assert.Equal(t, latestVersion.Version, test.latest)
		assert.Equal(t, isVersionEquel, test.current == test.latest)

		vulnraibilityList, err := c.GetVulnraibilityList(test.purl_name_With_cves)
		assert.NilError(t, err, "Get vulnrability list")
		assert.Equal(t, len(vulnraibilityList), test.vul_num)
	}
}

func TestPostBySbom(t *testing.T) {
	tests := []struct {
		fixture  string
		vers_num int
		vuln_num int
	}{
		{
			fixture:  "test-fixtures/sbom/python.sbom.json",
			vers_num: 5,
			vuln_num: 86,
		},
	}

	for _, test := range tests {
		t.Run(test.fixture, func(t *testing.T) {
			c, cyclonedx_manager, bom, _ := BasePostLogic(t, test.fixture)

			err := cyclonedx_manager.ReadFromFile(test.fixture, bom)
			assert.NilError(t, err, "Read from file")

			LatestVersionOfSbom, err := c.GetLatestVersionBySbom(bom)
			assert.NilError(t, err, "Get repository latest by sbom")
			count := 0
			for _, v := range LatestVersionOfSbom {
				if !v.IsVersionEquel {
					count += 1
				}
			}
			assert.Assert(t, count >= test.vers_num)

			vulnraibilityListOfSbom, err := c.GetVulnraibilityListBySbom(bom)
			assert.NilError(t, err, "Get vulnrability list by sbom")

			count = 0
			for _, v := range vulnraibilityListOfSbom {
				count += len(v)
			}
			assert.Assert(t, count >= test.vuln_num)
		})
	}
}
