package client

import (
	"encoding/json"
	"os"
	"time"
)

// WindrosePlus has no usable A2S responder and no public HTTP endpoint, so the
// API reads the dashboard's local status file directly. The file is owned by
// user `windrose` but world-readable (mode 0644) — the API user just needs
// filesystem access to /home/windrose/windrose/server-files/...
//
// Freshness gate matches windrose-metrics.sh: if the file hasn't been written
// in MaxAge, the server is considered offline (container restarting, WP+
// crashed, host hung).

type WindroseStatus struct {
	Server WindroseServerSection `json:"server"`
}

type WindroseServerSection struct {
	Name        string `json:"name"`
	PlayerCount int    `json:"player_count"`
	MaxPlayers  int    `json:"max_players"`
}

type WindroseClient struct {
	StatusPath string
	MaxAge     time.Duration
	Stat       func(string) (os.FileInfo, error)
	Read       func(string) ([]byte, error)
	Now        func() time.Time
}

func NewWindroseClient(statusPath string, maxAge time.Duration) WindroseClient {
	return WindroseClient{
		StatusPath: statusPath,
		MaxAge:     maxAge,
		Stat:       os.Stat,
		Read:       os.ReadFile,
		Now:        time.Now,
	}
}

// GetStatus returns the parsed status when the file exists and is fresh.
// A second return value of false means "treat as offline" — the caller does
// not need to distinguish missing / stale / unparseable.
func (c WindroseClient) GetStatus() (WindroseStatus, bool) {
	info, err := c.Stat(c.StatusPath)
	if err != nil {
		return WindroseStatus{}, false
	}
	if c.Now().Sub(info.ModTime()) > c.MaxAge {
		return WindroseStatus{}, false
	}
	data, err := c.Read(c.StatusPath)
	if err != nil {
		return WindroseStatus{}, false
	}
	var status WindroseStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return WindroseStatus{}, false
	}
	return status, true
}
