package assets

import "net/http"

func StartServer() {
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/", http.FileServer(AssetFile()))
		http.ListenAndServe(":8080", mux)
	}()
}
