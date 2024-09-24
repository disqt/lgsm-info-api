package model

import (
	"fmt"
	"lgsm-info-api/pkg/gameServers/model"
)

type ServerResponse interface {
	GetRunning() bool
	GetUrl() string
}

type OnlineServerResponse struct {
	Running    bool
	Url        string
	Redirect   string
	Players    int
	MaxPlayers int
}

func (o OnlineServerResponse) GetRunning() bool {
	return o.Running
}

func (o OnlineServerResponse) GetUrl() string {
	return o.Url
}

type OfflineServerResponse struct {
	Running bool
	Url     string
}

func (o OfflineServerResponse) GetRunning() bool {
	return o.Running
}

func (o OfflineServerResponse) GetUrl() string {
	return o.Url
}

func NewResponse(servers []model.GameServer) (map[string]ServerResponse, error) {
	response := make(map[string]ServerResponse)

	for _, server := range servers {
		switch v := server.(type) {
		case model.OnlineGameServer:
			url := ""
			if v.GetPort() == "" {
				url = v.GetHost()
			} else {
				url = v.GetHost() + ":" + v.GetPort()
			}

			response[server.GetName()] = OnlineServerResponse{
				Running:    v.GetIsOnline(),
				Url:        url,
				Redirect:   v.GetRedirect(),
				Players:    v.GetPlayers(),
				MaxPlayers: v.GetMaxPlayers(),
			}
		case model.OfflineGameServer:
			response[v.GetName()] = OfflineServerResponse{
				Running: v.GetIsOnline(),
				Url:     "",
			}
		default:
			fmt.Println("Unknown server type")
			return nil, fmt.Errorf("unknown server type")
		}
	}

	return response, nil
}
