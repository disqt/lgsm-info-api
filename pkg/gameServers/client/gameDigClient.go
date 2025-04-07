package client

import "os/exec"

type GameDigClient struct {
	GetServerInfo func(game string, host string, port string) ([]byte, error)
}

func NewGameDigClient() GameDigClient {
	return GameDigClient{
		GetServerInfo: ExecGameDigCommand,
	}
}

func ExecGameDigCommand(game string, host string, port string) ([]byte, error) {
	args := []string{"--type", game, host}

	if port != "" {
		args = append(args, "--port", port)
	}

	cmd := exec.Command("gamedig", args...)

	return cmd.Output()
}
