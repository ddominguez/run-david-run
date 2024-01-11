package utils

import (
	"fmt"
	"math"
	"net/url"
	"os"
)

// ActivityDistance returns an activity distance in miles
// formatted as 99.99 mi
func ActivityDistance(meters float64) string {
	d := meters * 0.000621371
	return fmt.Sprintf("%0.2f mi", d)
}

// ActivityPace returns the pace as minutes per mile: MM:SS /mi
func ActivityPace(meters float64, elapsedTime uint32) string {
	miles := meters * 0.000621371
	minutes := math.Floor(float64(elapsedTime / 60))
	pace := minutes / miles
	paceMinutes := math.Floor(pace)
	paceSeconds := math.Round((pace - paceMinutes) * 60)
	return fmt.Sprintf("%d:%02d /mi", int(paceMinutes), int(paceSeconds))
}

// TimeFormatted returns the elapsed time of the activity in the
// following format: HH:MM:SS
func TimeFormatted(elapsedTime uint32) string {
	hours := elapsedTime / 3600
	remainingSeconds := elapsedTime % 3600
	minutes := remainingSeconds / 60
	seconds := remainingSeconds % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

func getMapboxAcessToken() (string, error) {
	var token string
	var found bool
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "PRD" {
		token, found = os.LookupEnv("PRD_MAPBOX_ACCESS_TOKEN")
	} else {
		appEnv = "DEV"
		token, found = os.LookupEnv("DEV_MAPBOX_ACCESS_TOKEN")
	}
	if !found {
		return "", fmt.Errorf("%s mapbox access token not found", appEnv)
	}
	return token, nil
}

// MapboxURL returns a url to a static mapbox image
func MapboxURL(polyline string) string {
	token, err := getMapboxAcessToken()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	base := "https://api.mapbox.com/styles/v1/mapbox/streets-v12/static"
	params := fmt.Sprintf("logo=false&access_token=%s", token)
	escaped := url.QueryEscape(polyline)
	return fmt.Sprintf("%s/path-3+f11-0.6(%s)/auto/500x300?%s", base, escaped, params)
}
