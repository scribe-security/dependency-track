package main

import (
	"deptrack/client"
)

func main() {
	c, err := client.NewDepTrackClient("")
	if err != nil {
		panic(err)
	}
	err = c.Login("admin", "admin123")
	if err != nil {
		panic(err)
	}
	err = c.GetTeam()
	if err != nil {
		panic(err)
	}
	// client.NewTeam()
}
