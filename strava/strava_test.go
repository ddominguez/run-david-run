package strava

import (
	"net/url"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	token := "myToken"
	c := NewClient(token)

	expectedTimeout := 10 * time.Second
	resultTimeout := c.httpclient.Timeout
	if resultTimeout != expectedTimeout {
		t.Errorf("Incorrect client timeout value. Found(%d), Expected(%d)", resultTimeout, expectedTimeout)
	}

	if c.accessToken != token {
		t.Errorf("Incorrect access token. Found(%s), Expected(%s)", c.accessToken, token)
	}
}

func TestAuthorizationUrl(t *testing.T) {
	a := Authorization{
		ClientId:     "testid",
		ClientSecret: "testsecret",
		RedirectUri:  "http://test.com/redirect",
		Scope:        "testscope",
	}
	authUrl := a.Url()
	u, err := url.Parse(authUrl)
	if err != nil {
		t.Errorf("%s is an invalid url. %s", authUrl, err)
	}

	expectedPath := "/oauth/authorize"
	if u.Path != expectedPath {
		t.Errorf("Authorization url has incorrect path. Found(%s), Expected(%s)", u.Path, expectedPath)
	}

	testCases := []struct {
		expectedParam string
		expectedValue string
	}{
		{"client_id", a.ClientId},
		{"redirect_uri", a.RedirectUri},
		{"scope", a.Scope},
		{"response_type", "code"},
	}

	resParams := u.Query()
	for _, tc := range testCases {
		if !resParams.Has(tc.expectedParam) {
			t.Errorf("Authorization url expected `%s` in query params", tc.expectedParam)
		}
		if resParams.Get(tc.expectedParam) != tc.expectedValue {
			t.Errorf("Query param `%s` has incorrect value. Found(%s), Expected(%s)", tc.expectedParam, resParams.Get(tc.expectedParam), tc.expectedValue)
		}
	}
}

func TestNameSlugified(t *testing.T) {
	testCases := []struct {
		input    Activity
		expected string
	}{
		{Activity{Id: 1, Name: "NYC Marathon"}, "nyc-marathon"},
		{Activity{Id: 2, Name: "Test 5 & 10 Miler"}, "test-5-10-miler"},
		{Activity{Id: 3, Name: "St. Dave 10k"}, "st-dave-10k"},
	}

	for _, tc := range testCases {
		slug := tc.input.NameSlugified()
		if slug != tc.expected {
			t.Errorf("NameSlugified() has unexpected value. Found(%s), Expected(%s)", slug, tc.expected)
		}
	}
}
