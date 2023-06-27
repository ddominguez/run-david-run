package strava

import (
	"net/url"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	headers := make(map[string]string)
	c := NewClient(headers)

	expectedTimeout := 10 * time.Second
	resultTimeout := c.httpclient.Timeout
	if resultTimeout != expectedTimeout {
		t.Errorf("Incorrect client timeout value. Found(%d), Expected(%d)", resultTimeout, expectedTimeout)
	}

	expectedHeadersLen := len(headers)
	resultHeadersLen := len(c.headers)
	if resultHeadersLen != expectedHeadersLen {
		t.Errorf("Incorrect client headers length. Found(%d), Expected(%d)", resultHeadersLen, expectedHeadersLen)
	}
}

func TestNewClientWithHeaders(t *testing.T) {
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer Token"
	c := NewClient(headers)

	expectedHeaderName := "Authorization"
	val, ok := c.headers[expectedHeaderName]
	if !ok {
		t.Errorf("Expected header `%s` not found", expectedHeaderName)
	}

	expectedToken := "Bearer Token"
	if val != expectedToken {
		t.Errorf("Header `%s` has wrong value. Found(%s), Expected(%s)", expectedHeaderName, val, expectedToken)
	}

	expectedHeadersLen := len(headers)
	resultHeadersLen := len(c.headers)
	if resultHeadersLen != expectedHeadersLen {
		t.Errorf("Incorrect client headers length. Found(%d), Expected(%d)", resultHeadersLen, expectedHeadersLen)
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
