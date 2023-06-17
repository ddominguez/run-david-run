package strava

import (
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
