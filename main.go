package main

import (
	"github.com/dan-j/go-wasm/server"
	"log"
	"net/http"
	"strings"
)

const dir = "./html"

func main() {
	fs := http.FileServer(http.Dir(dir))
	log.Print("Serving " + dir + " on http://localhost:8080")
	hub := server.NewHub()
	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		log.Println("handling http req")
		resp.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(resp, req)
	})
	mux.HandleFunc("/ws/boards/1", func(w http.ResponseWriter, r *http.Request) {
		server.ServeWs(hub, w, r)
	})

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
