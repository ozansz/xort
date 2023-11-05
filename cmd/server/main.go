package main

import (
	"flag"
	"io"
	"log"
	"net/http"

	"go.sazak.io/xort"
)

var (
	port  = flag.String("port", xort.DefaultPort, "Port to listen on")
	users = flag.String("users", "users.json", "Path to users file")
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.LUTC)
	flag.Parse()

	authHandler, err := newFileAuthHandler(*users)
	if err != nil {
		log.Fatalf("Failed to initialize auth handler with the given users file %s: %v", *users, err)
	}

	sessionReg := xort.NewInMemorySessionRegistry()
	server := xort.NewServer(sessionReg, authHandler)

	server.Handle("echo", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, req.Body)
	})

	if http.ListenAndServe(*port, server) != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
