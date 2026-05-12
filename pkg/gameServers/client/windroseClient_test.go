package client

import (
	"errors"
	"os"
	"testing"
	"time"
)

type fakeFileInfo struct {
	os.FileInfo
	mtime time.Time
}

func (f fakeFileInfo) ModTime() time.Time { return f.mtime }

func newClient(stat func(string) (os.FileInfo, error), read func(string) ([]byte, error), now time.Time) WindroseClient {
	return WindroseClient{
		StatusPath: "/fake/path",
		MaxAge:     90 * time.Second,
		Stat:       stat,
		Read:       read,
		Now:        func() time.Time { return now },
	}
}

func TestGetStatus_Fresh(t *testing.T) {
	now := time.Unix(2_000_000_000, 0)
	c := newClient(
		func(string) (os.FileInfo, error) {
			return fakeFileInfo{mtime: now.Add(-10 * time.Second)}, nil
		},
		func(string) ([]byte, error) {
			return []byte(`{"server":{"name":"disqt.com","player_count":3,"max_players":10}}`), nil
		},
		now,
	)

	status, ok := c.GetStatus()
	if !ok {
		t.Fatal("expected ok=true for fresh file")
	}
	if status.Server.Name != "disqt.com" || status.Server.PlayerCount != 3 || status.Server.MaxPlayers != 10 {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestGetStatus_Stale(t *testing.T) {
	now := time.Unix(2_000_000_000, 0)
	c := newClient(
		func(string) (os.FileInfo, error) {
			return fakeFileInfo{mtime: now.Add(-5 * time.Minute)}, nil
		},
		func(string) ([]byte, error) { return []byte(`{}`), nil },
		now,
	)

	if _, ok := c.GetStatus(); ok {
		t.Fatal("expected ok=false for stale file")
	}
}

func TestGetStatus_Missing(t *testing.T) {
	c := newClient(
		func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		func(string) ([]byte, error) { return nil, nil },
		time.Now(),
	)

	if _, ok := c.GetStatus(); ok {
		t.Fatal("expected ok=false when stat fails")
	}
}

func TestGetStatus_ReadError(t *testing.T) {
	now := time.Now()
	c := newClient(
		func(string) (os.FileInfo, error) { return fakeFileInfo{mtime: now}, nil },
		func(string) ([]byte, error) { return nil, errors.New("read failed") },
		now,
	)

	if _, ok := c.GetStatus(); ok {
		t.Fatal("expected ok=false when read fails")
	}
}

func TestGetStatus_Malformed(t *testing.T) {
	now := time.Now()
	c := newClient(
		func(string) (os.FileInfo, error) { return fakeFileInfo{mtime: now}, nil },
		func(string) ([]byte, error) { return []byte(`not json`), nil },
		now,
	)

	if _, ok := c.GetStatus(); ok {
		t.Fatal("expected ok=false for malformed JSON")
	}
}
