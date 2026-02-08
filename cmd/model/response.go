package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lgsm-info-api/pkg/gameServers/model"
	"sort"
	"strings"
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

type serverEntry struct {
	name     string
	response ServerResponse
}

// OrderedServerMap is a JSON object with guaranteed key ordering.
type OrderedServerMap []serverEntry

func (m OrderedServerMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, entry := range m {
		if i > 0 {
			buf.WriteByte(',')
		}
		key, _ := json.Marshal(entry.name)
		val, err := json.Marshal(entry.response)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')
		buf.Write(val)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func NewResponse(servers []model.GameServer) (OrderedServerMap, error) {
	var entries []serverEntry

	for _, server := range servers {
		switch v := server.(type) {
		case model.OnlineGameServer:
			url := ""
			if v.GetPort() == "" {
				url = v.GetHost()
			} else {
				url = v.GetHost() + ":" + v.GetPort()
			}

			entries = append(entries, serverEntry{
				name: server.GetName(),
				response: OnlineServerResponse{
					Running:    v.GetIsOnline(),
					Url:        url,
					Redirect:   v.GetRedirect(),
					Players:    v.GetPlayers(),
					MaxPlayers: v.GetMaxPlayers(),
				},
			})
		case model.OfflineGameServer:
			entries = append(entries, serverEntry{
				name: v.GetName(),
				response: OfflineServerResponse{
					Running: v.GetIsOnline(),
					Url:     "",
				},
			})
		default:
			fmt.Println("Unknown server type")
			return nil, fmt.Errorf("unknown server type")
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		ri, rj := entries[i].response.GetRunning(), entries[j].response.GetRunning()
		if ri != rj {
			return ri
		}
		return strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
	})

	return OrderedServerMap(entries), nil
}
