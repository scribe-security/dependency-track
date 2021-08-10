package main

import (
	"database/sql"
	"deptrack/client"
	"deptrack/core"
	"deptrack/models"
	"fmt"
	"net/url"
	"os"

	// "github.com/jackc/pgx/v4"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	fixture := "test/integration/test-fixtures/sbom/python.sbom.json"
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

func main() {
	// Init db
	db := Connect()
	db.AutoMigrate(&models.SbomRequest{})
	sqldb, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()
	init_db(db)

	// Init managers
	c := initClient()
	cyclonedx_manager := initCycloneDxManager()

	// Read sbom from file
	bom, name := ReadSbom(cyclonedx_manager)

	// Post sbom to dep track - receive reponse
	sbom_response := PostSbom(name, c, bom)

	// Save data to DB
	sbom_req := models.SbomRequest{Status: "Pending", DepTrackSbomPostResponse: *sbom_response}
	err = models.CreateSbomRequest(&sbom_req)
	if err != nil {
		panic(err)
	}
	var find_SbomRequest models.SbomRequest
	err = models.GetSbomRequest(&find_SbomRequest, 1)
	if err != nil {
		panic(err)
	}

	fmt.Printf("FOUND: %+v\n", find_SbomRequest.DepTrackSbomPostResponse.Token)
}
