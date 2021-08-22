package main

import (

	// "github.com/jackc/pgx/v4"

	_ "gorm.io/driver/postgres"
)

func main() {
	// Test keys

	//api_key := "nssq2LQPmGBoWH8ixGwPhWdblxwcECZH"
	//os.Setenv("API_KEY", api_key)
	// purl := "pkg:pypi/argparse@1.2.1"
	// purl_with_cves := "pkg:deb/debian/git@1%3A2.20.1-2%20deb10u3?arch=amd64"
	//TestPostSbomAndCheck(api_key)
	// cyclonedx_manager := initCycloneDxManager()

	// // Read sbom from file
	// bom, name := ReadSbom(cyclonedx_manager)

	// // Init managers
	// client, err := client.NewDepTrackClient(api_key)
	// if err != nil {
	// 	panic(err)
	// }

	// component, err := client.GetComponentsByPURL(purl)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(component)

	// latest, err := client.GetRepositoryLatest(purl)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(latest)
	// latestVersion, currentVersion, isVersionEquel, err := client.GetLatestVersion(purl)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(latestVersion, currentVersion, isVersionEquel)

	// vulnraibilityList, err := client.GetVulnraibilityList(purl_with_cves)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(vulnraibilityList, currentVersion)

	// LatestVersionOfSbom, err := client.GetLatestVersionBySbom(bom)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(LatestVersionOfSbom, name)

	// client.GetVulnraibilityListBySbom(bom)
	// vulnraibilityListOfSbom, err := client.GetVulnraibilityListBySbom(bom)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(vulnraibilityListOfSbom, name)

	// // Init db
	// db := Connect()
	// db.AutoMigrate(&models.SbomRequest{})
	// sqldb, err := db.DB()
	// if err != nil {
	//      panic(err)
	// }
	// defer sqldb.Close()
	// init_db(db)

	// // Init managers
	// c := initClient()
	// cyclonedx_manager := initCycloneDxManager()

	// // Read sbom from file
	// bom, name := ReadSbom(cyclonedx_manager)

	// // Post sbom to dep track - receive reponse
	// sbom_response := PostSbom(name, c, bom)

	// // Save data to DB
	// sbom_req := models.SbomRequest{Status: "Pending", DepTrackSbomPostResponse: *sbom_response}
	// err = models.CreateSbomRequest(&sbom_req)
	// if err != nil {
	//      panic(err)
	// }
	// var find_SbomRequest models.SbomRequest
	// err = models.GetSbomRequest(&find_SbomRequest, 1)
	// if err != nil {
	//      panic(err)
	// }

	// fmt.Printf("FOUND: %+v\n", find_SbomRequest.DepTrackSbomPostResponse.Token)
}
