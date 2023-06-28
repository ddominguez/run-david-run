package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

const api_uri = "https://www.strava.com/api/v3"
const oauth_uri = "https://www.strava.com/oauth"

type Activity struct {
	Id             uint64    `json:"id"`
	Name           string    `json:"name"`
	Distance       float64   `json:"distance"`
	MovingTime     int       `json:"moving_time"`
	ElapsedTime    int       `json:"elapsed_time"`
	SportType      string    `json:"sport_type"`
	WorkoutType    uint8     `json:"workout_type"`
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

type Athlete struct {
	Id            uint64 `json:"id"`
	FirstName     string `json:"firstname"`
	LastName      string `json:"lastname"`
	Profile       string `json:"profile"`
	ProfileMedium string `json:"profile_medium"`
}

type authTokenResp struct {
	TokenType    string  `json:"token_type"`
	ExpiresAt    uint64  `json:"expires_at"`
	ExpiresIn    uint64  `json:"expires_in"`
	RefreshToken string  `json:"refresh_token"`
	AccessToken  string  `json:"access_token"`
	Athlete      Athlete `json:"athlete,omitempty"`
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
	ClientId     string
	ClientSecret string
	RedirectUri  string
	Scope        string
}

// Url returns a Url for authentication
func (a *Authorization) Url() string {
	qs := url.Values{}
	qs.Set("client_id", a.ClientId)
	qs.Set("redirect_uri", a.RedirectUri)
	qs.Set("response_type", "code")
	qs.Set("scope", a.Scope)
	return fmt.Sprintf("%s/authorize?%s", oauth_uri, qs.Encode())
}

func (a *Authorization) ReqAccessToken(code string) (authTokenResp, error) {
	payload := struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		GrantType    string `json:"grant_type"`
	}{
		ClientId:     a.ClientId,
		ClientSecret: a.ClientSecret,
		Code:         code,
		GrantType:    "authorization_code",
	}

	var tkResp authTokenResp

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return tkResp, fmt.Errorf("Unable to create request body for requesting access token. %s", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/token", oauth_uri), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return tkResp, fmt.Errorf("Unable to request access token. %s", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&tkResp); err != nil {
		return tkResp, fmt.Errorf("Failed to decode response body. %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return tkResp, fmt.Errorf(resp.Status)
	}
	return tkResp, nil
}

func (a *Authorization) RefreshToken(refreshToken string) (authTokenResp, error) {
	payload := map[string]string{
		"client_id":     a.ClientId,
		"client_secret": a.ClientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	var tkResp authTokenResp

	reqBody, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		return tkResp, nil
	}

	resp, err := http.Post(fmt.Sprintf("%s/token", oauth_uri), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return tkResp, fmt.Errorf("Unable to request access token. %s", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&tkResp); err != nil {
		return tkResp, fmt.Errorf("Failed to decode response body. %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return tkResp, fmt.Errorf(resp.Status)
	}

	return tkResp, nil
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
func GetActivities(c *Client, page uint16, perPage uint8) ([]Activity, error) {
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
