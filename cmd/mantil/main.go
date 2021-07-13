package main

import (
	"net/http"

	"github.com/atoz-technology/mantil-cli/cmd/mantil/cmd"
	"github.com/atoz-technology/mantil-cli/internal/assets"
)

func main() {
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/", http.FileServer(assets.AssetFile()))
		http.ListenAndServe(":8080", mux)
	}()
	cmd.Execute()
}
