package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
)

type SystemDirectory struct {
	Process string
}

func systemDirectory(w http.ResponseWriter, r *http.Request) {
	process, err := exec.Command("zsh", "-c", "ps aux | head -1; ps aux | grep ProjectZomboid64| sort -rnk 4 | more").Output()
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(SystemDirectory{Process: string(process)})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", systemDirectory)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
