package model

import (
	"encoding/json"
	"strings"
)

type GameDigResponse struct {
	Game       string
	Players    int         `json:"numplayers"`
	MaxPlayers int         `json:"maxplayers"`
	Port       json.Number `json:"queryPort"`
}

type GameServer interface {
	GetName() string
	GetIsOnline() bool
}

type OnlineGameServer struct {
	Name       string
	Host       string
	Port       string
	IsOnline   bool
	Players    int
	MaxPlayers int
	Redirect   string
}

func NewOnlineGameServer(name string, host string, port string, players int, maxPlayers int) OnlineGameServer {
	redirect := ""
	if strings.ToLower(name) == "minecraft" {
		port = ""
		redirect = "https://disqt.com/map/"
	}

	if strings.ToLower(name) == "xonotic" {
		redirect = "https://stats.xonotic.org/server/46827"
	}

	name = strings.ToUpper(string(name[0])) + name[1:]

	return OnlineGameServer{
		IsOnline:   true,
		Name:       name,
		Host:       host,
		Port:       port,
		Players:    players,
		MaxPlayers: maxPlayers,
		Redirect:   redirect,
	}
}

func (gameServer OnlineGameServer) GetIsOnline() bool {
	return gameServer.IsOnline
}

func (gameServer OnlineGameServer) GetName() string {
	return gameServer.Name
}

func (gameServer OnlineGameServer) GetHost() string {
	return gameServer.Host
}

func (gameServer OnlineGameServer) GetPort() string {
	return gameServer.Port
}

func (gameServer OnlineGameServer) GetPlayers() int {
	return gameServer.Players
}

func (gameServer OnlineGameServer) GetMaxPlayers() int {
	return gameServer.MaxPlayers
}

func (gameServer OnlineGameServer) GetRedirect() string {
	return gameServer.Redirect
}

type OfflineGameServer struct {
	Name     string
	IsOnline bool
}

func NewOfflineGameServer(name string) OfflineGameServer {
	name = strings.ToUpper(string(name[0])) + name[1:]
	return OfflineGameServer{
		Name:     name,
		IsOnline: false,
	}
}

func (gameServer OfflineGameServer) GetName() string {
	return gameServer.Name
}

func (gameServer OfflineGameServer) GetIsOnline() bool {
	return gameServer.IsOnline
}
