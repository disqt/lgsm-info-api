package client

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Palworld is not queryable through gamedig the way the other servers are.
// The dedicated server only starts its Steam A2S responder when launched with
// -publiclobby, and ours deliberately is not: it is a private server, and
// dropping the flag also keeps it off the Steam default query port, which is
// reserved for CS2 on this host. The query port is bound but silent.
//
// So status comes from Palworld's own REST API instead, which listens on
// localhost only. Two endpoints are used: /v1/api/metrics for the player
// counts, and /v1/api/info for the server name. Both require HTTP basic auth
// with the admin password from PalWorldSettings.ini.

type PalworldStatus struct {
	Name       string
	Players    int
	MaxPlayers int
}

type PalworldClient struct {
	BaseURL  string
	Username string
	Password string
	// Do is injectable so tests can serve canned responses without a listener.
	Do func(*http.Request) (*http.Response, error)
}

func NewPalworldClient(baseURL string, username string, password string, timeout time.Duration) PalworldClient {
	httpClient := &http.Client{Timeout: timeout}
	return PalworldClient{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		Do:       httpClient.Do,
	}
}

type palworldMetrics struct {
	CurrentPlayerNum int `json:"currentplayernum"`
	MaxPlayerNum     int `json:"maxplayernum"`
}

type palworldInfo struct {
	ServerName string `json:"servername"`
}

// GetStatus collapses unreachable / unauthorized / malformed into a single
// ok=false, so the caller treats every failure as "server offline" — the same
// contract as WindroseClient.GetStatus.
func (c PalworldClient) GetStatus() (PalworldStatus, bool) {
	// Without a password every request can only ever 401. Say offline rather
	// than hammering the REST API once per cache refresh.
	if c.Password == "" {
		return PalworldStatus{}, false
	}

	var metrics palworldMetrics
	if !c.get("/v1/api/metrics", &metrics) {
		return PalworldStatus{}, false
	}

	status := PalworldStatus{
		Players:    metrics.CurrentPlayerNum,
		MaxPlayers: metrics.MaxPlayerNum,
	}

	// A metrics response already proves the server is up, so a failing /info
	// must not downgrade it to offline. The name is cosmetic: it becomes Motd,
	// which is omitempty in the response.
	var info palworldInfo
	if c.get("/v1/api/info", &info) {
		status.Name = info.ServerName
	}

	return status, true
}

func (c PalworldClient) get(path string, out interface{}) bool {
	request, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return false
	}
	request.SetBasicAuth(c.Username, c.Password)
	request.Header.Set("Accept", "application/json")

	response, err := c.Do(request)
	if err != nil {
		return false
	}
	// Drain before closing: net/http only returns a connection to the idle pool
	// once its body is read to EOF. A wrong password means a permanent 401, so
	// without this the non-200 path would reconnect on every refresh forever.
	defer func() {
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return false
	}

	return json.NewDecoder(response.Body).Decode(out) == nil
}
