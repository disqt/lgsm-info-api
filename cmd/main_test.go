package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"lgsm-info-api/pkg/gameServers"
	"lgsm-info-api/pkg/gameServers/client"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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

		gameDigClientMock.On("GetServerInfo", "minecraft", "disqt.com", "").Return([]byte(`{"maxplayers":420,"numplayers":0,"queryPort": 25565,"players":[]}`), nil)
		gameDigClientMock.On("GetServerInfo", "valheim", "disqt.com", "").Return([]byte(`{"error":"Failed all 1 attempts"}`), nil)
		gameDigClientMock.On("GetServerInfo", "xonotic", "disqt.com", "26420").Return([]byte(`{"maxplayers":"420","numplayers":0,"queryPort": 26420,"players":[]}`), nil)
		gameDigClientMock.On("GetServerInfo", "csgo", "disqt.com", "27015").Return([]byte(`{"maxplayers":10,"numplayers":13,"queryPort": 27015,"players":[{"name":"Player1"},{"name":"Player2"},{"name":"Player3"}],"bots":[{"name":"Bot1"},{"name":"Bot2"},{"name":"Bot3"},{"name":"Bot4"},{"name":"Bot5"},{"name":"Bot6"},{"name":"Bot7"},{"name":"Bot8"},{"name":"Bot9"},{"name":"Bot10"}]}`), nil)

		gameDigClient := client.GameDigClient{
			GetServerInfo: gameDigClientMock.GetServerInfo,
		}

		cache := gameServers.NewServerCache(gameDigClient, 1*time.Hour)
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
				"Redirect": "https://disqt.com/map/"
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
}
