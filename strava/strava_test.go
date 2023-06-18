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
		t.Fatalf("Timeout result(%d), expected(%d)", resultTimeout, expectedTimeout)
	}

	expectedHeadersLen := len(headers)
	resultHeadersLen := len(c.headers)
	if resultHeadersLen != expectedHeadersLen {
		t.Fatalf("Headers length result(%d), expected(%d)", resultHeadersLen, expectedHeadersLen)
	}
}

func TestNewClientWithHeaders(t *testing.T) {
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer Token"
	c := NewClient(headers)

	val, ok := c.headers["Authorization"]
	if !ok {
		t.Fatal("Header `Authorization` not found")
	}

	expectedToken := "Bearer Token"
	if val != expectedToken {
		t.Fatalf("Header `Authorization` result(%s), expected(%s)", val, expectedToken)
	}

	expectedHeadersLen := len(headers)
	resultHeadersLen := len(c.headers)
	if resultHeadersLen != expectedHeadersLen {
		t.Fatalf("Headers length result(%d), expected(%d)", resultHeadersLen, expectedHeadersLen)
	}
}

func TestAuthorizationUrl(t *testing.T) {
	a := Authorization{
		clientId:     "testid",
		clientSecret: "testsecret",
		redirectUri:  "http://test.com/redirect",
		scope:        "testscope",
	}
	authUrl := a.url()
	u, err := url.Parse(authUrl)
	if err != nil {
		t.Fatalf("%s is an invalid url", authUrl)
	}

	expectedPath := "/oauth/authorize"
	if u.Path != expectedPath {
		t.Fatalf("Authorization url has incorrect path. Found(%s), Expected(%s)", u.Path, expectedPath)
	}

	testCases := []struct {
		expectedParam string
		expectedValue string
	}{
		{"client_id", a.clientId},
		{"redirect_uri", a.redirectUri},
		{"scope", a.scope},
		{"response_type", "code"},
	}

	resParams := u.Query()
	for _, tc := range testCases {
		if !resParams.Has(tc.expectedParam) {
			t.Fatalf("Authorization url expected `%s` in query params", tc.expectedParam)
		}
		if resParams.Get(tc.expectedParam) != tc.expectedValue {
			t.Fatalf("Query param `%s` result(%s). expected(%s)", tc.expectedParam, resParams.Get(tc.expectedParam), tc.expectedValue)
		}
	}
}
