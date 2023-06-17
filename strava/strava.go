package strava

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const api_uri = "https://www.strava.com/api/v3"
const oauth_uri = "https://www.strava.com/oauth"

type Activity struct {
	Id             int64     `json:"id"`
	Name           string    `json:"name"`
	Distance       float64   `json:"distance"`
	MovingTime     int       `json:"moving_time"`
	ElapsedTime    int       `json:"elapsed_time"`
	Type           string    `json:"type"`
	StartDate      time.Time `json:"start_date"`
	StartDateLocal time.Time `json:"start_date_local"`
	TimeZone       string    `json:"time_zone"`
	City           string    `json:"location_city"`
	State          string    `json:"location_state"`
	Country        string    `json:"location_country"`
	PhotoCount     int       `json:"photo_count"`
	Map            struct {
		Id              string `json:"id"`
		Polyline        string `json:"polyline"`
		SummaryPolyline string `json:"summary_polyline"`
	} `json:"map"`
}

type Client struct {
	httpclient http.Client
	headers    map[string]string
}

func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}
	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}
	return resp, nil
}

func NewClient(headers map[string]string) *Client {
	return &Client{
		httpclient: http.Client{Timeout: 10 * time.Second},
		headers:    headers,
	}
}

type Authorization struct {
	clientId     string
	clientSecret string
	redirectUri  string
	scope        string
}

// url returns a url for authentication
func (a *Authorization) url() string {
	qs := url.Values{}
	qs.Set("client_id", a.clientId)
	qs.Set("redirect_uri", a.redirectUri)
	qs.Set("response_type", "code")
	qs.Set("scope", a.scope)
	return fmt.Sprintf("%s?%s", oauth_uri, qs.Encode())
}

// GetActivity will return a strava activity for an authorized user
func GetActivity(c *Client, id uint64) (Activity, error) {
	var activity Activity

	resp, err := c.Get(fmt.Sprintf("%s/activities/%d", api_uri, id))
	if err != nil {
		return activity, err
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		return activity, err
	}
	return activity, nil
}

// GetActivities will return an array of strava activities for an authorized user
func GetActivities(c *Client, page int8, perPage int8) ([]Activity, error) {
	var activities []Activity

	qs := url.Values{}
	qs.Set("page", fmt.Sprintf("%d", page))
	qs.Set("per_page", fmt.Sprintf("%d", perPage))

	resp, err := c.Get(fmt.Sprintf("%s/activities?%s", api_uri, qs.Encode()))
	if err != nil {
		return activities, err
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return activities, err
	}
	return activities, nil
}
