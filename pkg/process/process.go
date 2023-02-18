package process

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
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
		// Split the output by new line and get the first result
		process := strings.Split(string(out), "\n")[0]
		// split the output by space
		col := strings.Split(process, " ")
		return Process{
			Pid: col[0],
			Out: process,
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
