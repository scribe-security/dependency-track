package integration

import (
	"deptrack/client"
	"os"
	"testing"
)

func GetLocalDepClient(t *testing.T) *client.DepTrackClient {
	api_key, ok := os.LookupEnv("API_KEY")
	if !ok {
		t.Fatalf("No api key")
	}
	c, err := client.NewDepTrackClient(api_key)
	if err != nil {
		t.Fatalf("Failed to create client, Err: %+v", err)
	}

	return c

}
