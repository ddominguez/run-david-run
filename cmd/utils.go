package cmd

import (
	"fmt"
	"net/url"
	"os"
)

func mapboxURL(polyline string) string {
	token, found := os.LookupEnv("MAPBOX_ACCESS_TOKEN")
	if !found {
		fmt.Println("MAPBOX_ACCESS_TOKEN is not set")
		return ""
	}

	base := "https://api.mapbox.com/styles/v1/mapbox/streets-v12/static"
	params := fmt.Sprintf("logo=false&access_token=%s", token)
	escaped := url.QueryEscape(polyline)
	return fmt.Sprintf("%s/path-3+f11-0.6(%s)/auto/500x300?%s", base, escaped, params)
}
