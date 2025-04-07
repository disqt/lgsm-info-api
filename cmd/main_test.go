package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"lgsm-info-api/pkg/gameServers/client"
	"net/http"
	"net/http/httptest"
	"testing"
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

	t.Run("Minecraft On, Valheim Off", func(t *testing.T) {
		gameDigClientMock := new(MockedGameDigClient)

		// Set up mock behavior
		gameDigClientMock.On("GetServerInfo", "minecraft", "disqt.com", "").Return([]byte(`{"maxplayers":420,"numplayers":0,"queryPort": 25565}`), nil)
		gameDigClientMock.On("GetServerInfo", "valheim", "disqt.com", "").Return([]byte(`{"error":"Failed all 1 attempts"}`), nil)
		gameDigClientMock.On("GetServerInfo", "xonotic", "disqt.com", "26420").Return([]byte(`{"maxplayers":"420","numplayers":0,"queryPort": 26420}`), nil)

		// Inject mock into the router
		gameDigClient := client.GameDigClient{
			GetServerInfo: gameDigClientMock.GetServerInfo,
		}
		r := setupRouter(gameDigClient)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/servers", nil)
		r.ServeHTTP(w, req)

		// Assert the result
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Assert that the mock was called
		gameDigClientMock.AssertExpectations(t)

		// Assert that the body matches
		expectedBody := `{
			"Minecraft": {
				"Url": "disqt.com",
				"Running": true,
				"Players": 0,
				"MaxPlayers": 420,
				"Redirect": "https://disqt.com/map/"
			},
			"Valheim": {
				"Url": "",
				"Running": false
			},
			"Xonotic": {
				"Url": "disqt.com:26420",
				"Running": true,
				"Players": 0,
				"MaxPlayers": 420,	
				"Redirect": "https://stats.xonotic.org/server/46827"
			}
		}`
		assert.JSONEq(t, expectedBody, w.Body.String())
	})
}
