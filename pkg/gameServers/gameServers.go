package gameServers

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type process struct {
	Pid string
	Out string
}

type Server struct {
	Running bool
	Pid     string
}

var servers = map[string]string{
	"Zomboid":   "Zomboid.*servername|servername.Zomboid",
	"Minecraft": "java -Xms2560M",
	"Valheim":   "Valheim.*servername|servername.Valheim",
}

func getRunningProcessByRegex(name string) (process, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("pgrep -f '%s'", name))
	log.Default().Println("Running command: " + cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Return an error
		return process{
			Out: string(out),
		}, err
	} else {
		// Split the output by new line and get the first result
		p := strings.Split(string(out), "\n")[0]
		// split the output by space
		col := strings.Split(p, " ")
		return process{
			Pid: col[0],
			Out: p,
		}, nil
	}
}

func GetRunningGameServers() map[string]Server {
	res := make(map[string]Server, len(servers))

	for serverName, regex := range servers {
		out, err := getRunningProcessByRegex(regex)
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
