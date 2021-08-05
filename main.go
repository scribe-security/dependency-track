package main

import (
	"deptrack/client"
	"fmt"
	"os"
)

func main() {
	api_key, ok := os.LookupEnv("API_KEY")
	if !ok {
		panic("No api key")
	}
	c, err := client.NewDepTrackClient(api_key)
	if err != nil {
		panic(err)
	}

	dst, err := c.GetTeam()
	if err != nil {
		panic(err)
	}
	fmt.Println(dst)

}
