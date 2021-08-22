package integration

import (
	"database/sql"
	"deptrack/client"
	"deptrack/core"
	"deptrack/models"
	"fmt"
	"net/url"
	"os"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gotest.tools/assert"
)

const SbomLocalPath string = "test/integration/test-fixtures/sbom/python.sbom.json"

func Connect() *gorm.DB {
	dsn := url.URL{
		User:     url.UserPassword("client", "client"),
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%d", "localhost", 25432),
		Path:     "client",
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	conn, err := sql.Open("pgx", dsn.String())
	if err != nil {
		panic(err.Error())
	}

	config := postgres.Config{
		DSN:                  dsn.String(),
		PreferSimpleProtocol: true,
		Conn:                 conn,
	}

	dialactor := postgres.New(config)
	db, err := gorm.Open(dialactor, &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}

	return db
}

func init_db(db *gorm.DB) {
	models.DB = db
}

func initClient() *client.DepTrackClient {
	api_key, ok := os.LookupEnv("API_KEY")
	if !ok {
		panic("No api key")
	}
	c, err := client.NewDepTrackClient(api_key)
	if err != nil {
		panic(err)
	}

	return c
}

func initCycloneDxManager() *core.CycloneDxManager {
	cyclonedx_manager, err := core.NewCycloneDxManager(core.JSON_FORMAT)
	if err != nil {
		panic(err)
	}
	return cyclonedx_manager
}

func ReadSbom(m *core.CycloneDxManager) (*cdx.BOM, string) {
	fixture := SbomLocalPath
	var bom cdx.BOM
	err := m.ReadFromFile(fixture, &bom)
	if err != nil {
		panic(err)
	}

	name, err := m.GetName(&bom)
	if err != nil {
		panic(err)
	}

	return &bom, name
}

func PostSbom(name string, c *client.DepTrackClient, bom *cdx.BOM) *client.DepTrackSbomPostResponse {
	params := client.DepTrackSbomPost{
		AutoCreate:  "true",
		ProjectName: name,
	}
	var sbom_response client.DepTrackSbomPostResponse
	err := c.PostSbom("bom", &params, bom, &sbom_response)
	if err != nil {
		panic(err)
	}

	return &sbom_response
}

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

func TestPostSbomAndCheck(t *testing.T) {

	api_key := "nssq2LQPmGBoWH8ixGwPhWdblxwcECZH"
	os.Setenv("API_KEY", api_key)
	// Init managers
	client := initClient()
	cyclonedx_manager := initCycloneDxManager()

	// Read sbom from file
	bom, name := ReadSbom(cyclonedx_manager)

	// Post sbom to dep track - receive reponse
	sbom_response := PostSbom(name, client, bom)

	is_finish_upload, err := client.WaitforSbomFinishUpload(sbom_response.Token)
	if err != nil {
		panic(err)
	}
	fmt.Println(sbom_response)
	fmt.Println(is_finish_upload)
}
