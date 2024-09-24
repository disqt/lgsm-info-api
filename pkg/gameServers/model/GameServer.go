package model

import "strings"

type GameDigResponse struct {
	Game       string
	Players    int    `json:"numplayers"`
	MaxPlayers int    `json:"maxplayers"`
	Port       string `json:"port"`
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
}

func (gameServer OnlineGameServer) GetIsOnline() bool {
	return true
}

func (gameServer OnlineGameServer) GetName() string {
	return strings.ToUpper(string(gameServer.Name[0])) + gameServer.Name[1:]
}

func (gameServer OnlineGameServer) GetHost() string {
	return gameServer.Host
}

func (gameServer OnlineGameServer) GetPort(gameDigResponse GameDigResponse) string {
	if gameDigResponse.Game == "minecraft" {
		return "" // This is because minecraft clients already assume the port is 25565
	}
	return gameServer.Port
}

func (gameServer OnlineGameServer) GetPlayers() int {
	return gameServer.Players
}

func (gameServer OnlineGameServer) GetMaxPlayers() int {
	return gameServer.MaxPlayers
}

type OfflineGameServer struct {
	Name     string
	IsOnline bool
}

func (gameServer OfflineGameServer) GetName() string {
	return gameServer.Name
}

func (gameServer OfflineGameServer) GetIsOnline() bool {
	return false
}
