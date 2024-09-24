package client

import "os/exec"

type GameDigClient struct {
	GetServerInfo func(game string, host string) ([]byte, error)
}

func NewGameDigClient() GameDigClient {
	return GameDigClient{
		GetServerInfo: ExecGameDigCommand,
	}
}

func ExecGameDigCommand(game string, host string) ([]byte, error) {
	cmd := exec.Command("gamedig", "--type", game, host)

	return cmd.Output()
}
