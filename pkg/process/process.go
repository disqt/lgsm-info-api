package process

import (
	"fmt"
	"log"
	"os/exec"
)

type Process struct {
	Pid string
	Out string
}

type Server struct {
	Running bool
	Pid     string
}

const (
	Zomboid   string = "Zomboid"
	Minecraft        = "Minecraft"
	Valheim          = "Valheim"
)

func getRunningProcessByName(name string) (Process, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("pgrep -f '%s'", name))
	log.Default().Println("Running command: " + cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Return an error
		return Process{
			Out: string(out),
		}, err
	} else {
		return Process{
			Pid: string(out),
			Out: string(out),
		}, nil
	}
}

func GetRunningServers() map[string]Server {
	serverNames := []string{Zomboid, Minecraft, Valheim}
	res := make(map[string]Server, len(serverNames))

	for _, serverName := range serverNames {
		regex := fmt.Sprintf("%s.*servername|servername.*%s", serverName, serverName)
		out, err := getRunningProcessByName(regex)
		if err != nil {
			log.Default().Println(fmt.Sprint(err) + ": " + out.Out)
			res[serverName] = Server{
				Running: false,
				Pid:     "",
			}
		} else {
			res[serverName] = Server{
				Running: true,
				Pid:     out.Pid,
			}
		}
	}

	return res
}
