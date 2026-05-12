package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"lgsm-info-api/pkg/gameServers"
	"lgsm-info-api/pkg/gameServers/client"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type fakeFileInfo struct {
	os.FileInfo
	mtime time.Time
}

func (f fakeFileInfo) ModTime() time.Time { return f.mtime }

func freshWindroseClient(body []byte) client.WindroseClient {
	now := time.Unix(2_000_000_000, 0)
	return client.WindroseClient{
		StatusPath: "/fake",
		MaxAge:     90 * time.Second,
		Stat:       func(string) (os.FileInfo, error) { return fakeFileInfo{mtime: now}, nil },
		Read:       func(string) ([]byte, error) { return body, nil },
		Now:        func() time.Time { return now },
	}
}

func offlineWindroseClient() client.WindroseClient {
	return client.WindroseClient{
		StatusPath: "/fake",
		MaxAge:     90 * time.Second,
		Stat:       func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		Read:       func(string) ([]byte, error) { return nil, nil },
		Now:        time.Now,
	}
}

type MockedGameDigClient struct {
	mock.Mock
}

func (m *MockedGameDigClient) GetServerInfo(game string, host string, port string) ([]byte, error) {
	args := m.Called(game, host, port)
	return args.Get(0).([]byte), args.Error(1)
}

func TestGetServersHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Minecraft On, Valheim Off, Xonotic On, CS2 On", func(t *testing.T) {
		gameDigClientMock := new(MockedGameDigClient)

		gameDigClientMock.On("GetServerInfo", "minecraft", "disqt.com", "").Return([]byte(`{"name":"DISQT Minecraft","maxplayers":420,"numplayers":0,"queryPort": 25565,"players":[]}`), nil)
		gameDigClientMock.On("GetServerInfo", "valheim", "disqt.com", "").Return([]byte(`{"error":"Failed all 1 attempts"}`), nil)
		gameDigClientMock.On("GetServerInfo", "xonotic", "disqt.com", "26420").Return([]byte(`{"maxplayers":"420","numplayers":0,"queryPort": 26420,"players":[]}`), nil)
		gameDigClientMock.On("GetServerInfo", "csgo", "disqt.com", "27015").Return([]byte(`{"maxplayers":10,"numplayers":13,"queryPort": 27015,"players":[{"name":"Player1"},{"name":"Player2"},{"name":"Player3"}],"bots":[{"name":"Bot1"},{"name":"Bot2"},{"name":"Bot3"},{"name":"Bot4"},{"name":"Bot5"},{"name":"Bot6"},{"name":"Bot7"},{"name":"Bot8"},{"name":"Bot9"},{"name":"Bot10"}]}`), nil)

		gameDigClient := client.GameDigClient{
			GetServerInfo: gameDigClientMock.GetServerInfo,
		}

		windroseClient := freshWindroseClient([]byte(`{"server":{"name":"disqt.com","player_count":2,"max_players":10}}`))

		cache := gameServers.NewServerCache(gameDigClient, windroseClient, 1*time.Hour)
		cache.Start()

		r := setupRouter(cache)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/servers", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		gameDigClientMock.AssertExpectations(t)

		expectedBody := `{
			"Counter Strike 2": {
				"Url": "disqt.com",
				"Running": true,
				"Players": 3,
				"MaxPlayers": 10,
				"Redirect": "steam://rungameid/730//+connect disqt.com:27015"
			},
			"Minecraft": {
				"Url": "disqt.com",
				"Running": true,
				"Players": 0,
				"MaxPlayers": 420,
				"Redirect": "https://disqt.com/minecraft",
				"Motd": "DISQT Minecraft"
			},
			"Windrose": {
				"Url": "",
				"Running": true,
				"Players": 2,
				"MaxPlayers": 10,
				"Redirect": "",
				"Motd": "disqt.com"
			},
			"Xonotic": {
				"Url": "disqt.com:26420",
				"Running": true,
				"Players": 0,
				"MaxPlayers": 420,
				"Redirect": "https://stats.xonotic.org/server/46827"
			},
			"Valheim": {
				"Url": "",
				"Running": false
			}
		}`
		assert.JSONEq(t, expectedBody, w.Body.String())
	})

	t.Run("Windrose offline when status file missing", func(t *testing.T) {
		gameDigClientMock := new(MockedGameDigClient)
		gameDigClientMock.On("GetServerInfo", "minecraft", "disqt.com", "").Return([]byte(`{"error":"x"}`), nil)
		gameDigClientMock.On("GetServerInfo", "valheim", "disqt.com", "").Return([]byte(`{"error":"x"}`), nil)
		gameDigClientMock.On("GetServerInfo", "xonotic", "disqt.com", "26420").Return([]byte(`{"error":"x"}`), nil)
		gameDigClientMock.On("GetServerInfo", "csgo", "disqt.com", "27015").Return([]byte(`{"error":"x"}`), nil)

		gameDigClient := client.GameDigClient{GetServerInfo: gameDigClientMock.GetServerInfo}
		cache := gameServers.NewServerCache(gameDigClient, offlineWindroseClient(), 1*time.Hour)
		cache.Start()

		r := setupRouter(cache)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/servers", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"Windrose": {`)
		assert.Contains(t, w.Body.String(), `"Running": false`)
	})
}
