package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const api_uri = "https://www.strava.com/api/v3"
const oauth_uri = "https://www.strava.com/oauth"

type Activity struct {
	Id             uint64  `json:"id"`
	Name           string  `json:"name"`
	Distance       float64 `json:"distance"`
	MovingTime     uint32  `json:"moving_time"`
	ElapsedTime    uint32  `json:"elapsed_time"`
	SportType      string  `json:"sport_type"`
	WorkoutType    uint8   `json:"workout_type"`
	StartDateLocal string  `json:"start_date_local"`
	Map            struct {
		Id              string `json:"id"`
		SummaryPolyline string `json:"summary_polyline"`
	} `json:"map"`
}

func (a *Activity) DistanceInMiles() string {
	d := a.Distance * 0.000621371
	return fmt.Sprintf("%0.2f mi", d)
}

func (a *Activity) Pace() string {
	miles := a.Distance * 0.000621371
	minutes := math.Floor(float64(a.ElapsedTime / 60))
	pace := minutes / miles
	paceMinutes := math.Floor(pace)
	paceSeconds := math.Round((pace - paceMinutes) * 60)
	return fmt.Sprintf("%d:%02d /mi", int(paceMinutes), int(paceSeconds))
}

func (a *Activity) TimeFormatted() string {
	hours := a.ElapsedTime / 3600
	remainingSeconds := a.ElapsedTime % 3600
	minutes := remainingSeconds / 60
	seconds := remainingSeconds % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

// IsRace will return true for running race events
func (a *Activity) IsRace() bool {
	return a.SportType == "Run" && a.WorkoutType == 1
}

var re = regexp.MustCompile("[^a-z0-9]+")

func (a *Activity) NameSlugified() string {
	return strings.Trim(re.ReplaceAllString(strings.ToLower(a.Name), "-"), "-")
}

type Athlete struct {
	Id            uint64 `json:"id"`
	FirstName     string `json:"firstname"`
	LastName      string `json:"lastname"`
	Profile       string `json:"profile"`
	ProfileMedium string `json:"profile_medium"`
}

type AuthTokenResp struct {
	TokenType    string  `json:"token_type"`
	ExpiresAt    uint64  `json:"expires_at"`
	ExpiresIn    uint64  `json:"expires_in"`
	RefreshToken string  `json:"refresh_token"`
	AccessToken  string  `json:"access_token"`
	Athlete      Athlete `json:"athlete,omitempty"`
}

type Client struct {
	httpclient  http.Client
	accessToken string
}

func (c *Client) Get(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
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

// NewClient creates and returns a new Client
// which contains an httpclient with a specified Timeout
// and an accessToken for strava api requests
func NewClient(t string) *Client {
	return &Client{
		httpclient:  http.Client{Timeout: 10 * time.Second},
		accessToken: t,
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

func (a *Authorization) ReqAccessToken(code string) (AuthTokenResp, error) {
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

	var tkResp AuthTokenResp

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

func (a *Authorization) RefreshToken(refreshToken string) (AuthTokenResp, error) {
	payload := map[string]string{
		"client_id":     a.ClientId,
		"client_secret": a.ClientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	var tkResp AuthTokenResp

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

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.accessToken),
	}
	resp, err := c.Get(fmt.Sprintf("%s/activities/%d", api_uri, id), headers)
	if err != nil {
		return activity, err
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		return activity, err
	}
	return activity, nil
}

type ReqParams struct {
	Page    uint16
	PerPage uint8
	After   int64
}

// GetActivities will return an array of strava activities for an authorized user
func GetActivities(c *Client, rp ReqParams) ([]Activity, error) {
	var activities []Activity

	qs := url.Values{}
	qs.Set("page", fmt.Sprintf("%d", rp.Page))
	qs.Set("per_page", fmt.Sprintf("%d", rp.PerPage))
	qs.Set("after", fmt.Sprintf("%d", rp.After))

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.accessToken),
	}
	url := fmt.Sprintf("%s/athlete/activities?%s", api_uri, qs.Encode())
	resp, err := c.Get(url, headers)
	if err != nil {
		return activities, err
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return activities, err
	}
	return activities, nil
}
