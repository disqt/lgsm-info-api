package client

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// respond builds a canned http.Response for the fake transport.
func respond(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// routed dispatches on URL path so a single fake can serve both endpoints.
func routed(bodies map[string]*http.Response, err error) func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		if err != nil {
			return nil, err
		}
		if resp, ok := bodies[req.URL.Path]; ok {
			return resp, nil
		}
		return respond(404, `{}`), nil
	}
}

func newPalworldClient(do func(*http.Request) (*http.Response, error)) PalworldClient {
	return PalworldClient{
		BaseURL:  "http://127.0.0.1:8212",
		Username: "admin",
		Password: "secret",
		Do:       do,
	}
}

const (
	metricsBody = `{"currentplayernum":3,"serverfps":59,"maxplayernum":8,"uptime":176}`
	infoBody    = `{"version":"v1.0.5.102999","servername":"disqt Palworld"}`
)

func TestPalworldGetStatus_Online(t *testing.T) {
	c := newPalworldClient(routed(map[string]*http.Response{
		"/v1/api/metrics": respond(200, metricsBody),
		"/v1/api/info":    respond(200, infoBody),
	}, nil))

	status, ok := c.GetStatus()
	if !ok {
		t.Fatal("expected ok=true when both endpoints answer")
	}
	if status.Players != 3 || status.MaxPlayers != 8 {
		t.Fatalf("unexpected player counts: %+v", status)
	}
	if status.Name != "disqt Palworld" {
		t.Fatalf("unexpected server name: %q", status.Name)
	}
}

// The metrics call alone proves the server is up. A failing /info must not
// downgrade a live server to offline — the name is cosmetic (it becomes Motd,
// which is omitempty in the response).
func TestPalworldGetStatus_OnlineWithoutName(t *testing.T) {
	c := newPalworldClient(routed(map[string]*http.Response{
		"/v1/api/metrics": respond(200, metricsBody),
		"/v1/api/info":    respond(500, `boom`),
	}, nil))

	status, ok := c.GetStatus()
	if !ok {
		t.Fatal("expected ok=true when metrics answers but info fails")
	}
	if status.Players != 3 || status.MaxPlayers != 8 {
		t.Fatalf("unexpected player counts: %+v", status)
	}
	if status.Name != "" {
		t.Fatalf("expected empty name, got %q", status.Name)
	}
}

func TestPalworldGetStatus_Unreachable(t *testing.T) {
	c := newPalworldClient(routed(nil, errors.New("connection refused")))

	if _, ok := c.GetStatus(); ok {
		t.Fatal("expected ok=false when the transport fails")
	}
}

// A wrong admin password must read as offline rather than crashing the refresh.
func TestPalworldGetStatus_Unauthorized(t *testing.T) {
	c := newPalworldClient(routed(map[string]*http.Response{
		"/v1/api/metrics": respond(401, `{"error":"unauthorized"}`),
	}, nil))

	if _, ok := c.GetStatus(); ok {
		t.Fatal("expected ok=false on 401")
	}
}

func TestPalworldGetStatus_Malformed(t *testing.T) {
	c := newPalworldClient(routed(map[string]*http.Response{
		"/v1/api/metrics": respond(200, `not json`),
	}, nil))

	if _, ok := c.GetStatus(); ok {
		t.Fatal("expected ok=false for malformed JSON")
	}
}

// With no password configured the client must stay quiet rather than spamming
// the REST API with requests that can only ever 401.
func TestPalworldGetStatus_NoPasswordConfigured(t *testing.T) {
	called := false
	c := newPalworldClient(func(*http.Request) (*http.Response, error) {
		called = true
		return respond(200, metricsBody), nil
	})
	c.Password = ""

	if _, ok := c.GetStatus(); ok {
		t.Fatal("expected ok=false when no password is configured")
	}
	if called {
		t.Fatal("expected no HTTP call when no password is configured")
	}
}

func TestPalworldGetStatus_SendsBasicAuth(t *testing.T) {
	var gotUser, gotPass string
	var gotOK bool
	c := newPalworldClient(func(req *http.Request) (*http.Response, error) {
		gotUser, gotPass, gotOK = req.BasicAuth()
		return respond(200, metricsBody), nil
	})

	c.GetStatus()

	if !gotOK {
		t.Fatal("expected basic auth on the request")
	}
	if gotUser != "admin" || gotPass != "secret" {
		t.Fatalf("unexpected credentials: %q / %q", gotUser, gotPass)
	}
}
