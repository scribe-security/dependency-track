package main

func main() {
	// api_key, ok := os.LookupEnv("API_KEY")
	// if !ok {
	// 	panic("No api key")
	// }
	// c, err := client.NewDepTrackClient(api_key)
	// if err != nil {
	// 	panic(err)
	// }

	// dst, err := c.GetTeam()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(dst)

	// var bom cdx.BOM
	// cyclonedx_manager, err := core.NewCycloneDxManager(core.JSON_FORMAT)
	// if err != nil {
	// 	panic(err)
	// }
	// sbom_path := "test/integration/test-fixtures/sbom/python.sbom.json"
	// err = cyclonedx_manager.ReadFromFile(sbom_path, &bom)
	// if err != nil {
	// 	panic(err)
	// }

	// name, err := cyclonedx_manager.GetName(&bom)
	// if err != nil {
	// 	panic(err)
	// }

	// params := client.DepTrackSbomPost{
	// 	AutoCreate:  "true",
	// 	ProjectName: name,
	// }

	// var sbom_response client.DepTrackSbomPostResponse
	// err = c.PostSbom("bom", &params, &bom, &sbom_response)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("SBOM token:", sbom_response)
}
