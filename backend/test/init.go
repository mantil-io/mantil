package test

import (
	"log"
	"os"
	"os/exec"
)

var apiURL = ""

func init() {
	if val, ok := os.LookupEnv("MANTIL_API_URL"); ok {
		apiURL = val
		return
	}
	out, err := exec.Command("mantil", "env", "-u").Output()
	if err != nil {
		log.Fatalf("can't find api url, execute of `mantil env -u` failed %v", err)
	}
	apiURL = string(out)
}
