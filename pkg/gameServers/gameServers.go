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

type server struct {
	Url   string
	Regex string
}

var servers = map[string]server{
	"Zomboid": {
		Url:   "http://disqt.com:16261/",
		Regex: "Zomboid.*servername|servername.Zomboid",
	},
	"Minecraft": {
		Url:   "http://disqt.com/",
		Regex: "java -Xms2560M",
	},
	"Valheim": {
		Url:   "http://disqt.com:2456/",
		Regex: "Valheim.*servername|servername.Valheim",
	},
}

type Server struct {
	Running bool
	Url     string
}

func getRunningProcessByRegex(regex string) (process, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("pgrep -f '%s'", regex))
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

	for serverName, server := range servers {
		out, err := getRunningProcessByRegex(server.Regex)
		if err != nil {
			log.Default().Println(fmt.Sprint(err) + ": " + out.Out)
			res[serverName] = Server{
				Running: false,
				Url:     server.Url,
			}
		} else {
			// Here, we can use out.Pid to fetch more information about that process
			res[serverName] = Server{
				Running: true,
				Url:     server.Url,
			}
		}
	}

	return res
}
