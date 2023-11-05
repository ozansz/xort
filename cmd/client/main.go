package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"go.sazak.io/xort"
)

var (
	host     = flag.String("host", "localhost", "Host to connect to")
	port     = flag.String("port", xort.DefaultPort, "Port to connect to")
	path     = flag.String("path", "", "Path to send request to")
	body     = flag.String("body", "", "Optional JSON body to send with the request")
	username = flag.String("username", "", "Username to login with")
	password = flag.String("password", "", "Password to login with")
	useTLS   = flag.Bool("tls", false, "Use TLS to connect to the server")
)

func main() {
	log.SetPrefix("xort: ")
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatalf("Username and password must be provided to login.")
	}
	if *path == "" {
		log.Fatalf("Path must be provided to send a request to.")
	}

	schema := "http"
	if *useTLS {
		schema = "https"
	}

	ctx := context.Background()
	client := xort.NewClient(fmt.Sprintf("%s://%s%s", schema, *host, *port))

	if err := client.Login(ctx, &xort.UserCredentials{
		Username: *username,
		Password: *password,
	}); err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	log.Printf("Successfully logged in.")

	resp, err := client.RawRequest(ctx, *path, []byte(*body))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	log.Printf("Response: %s", string(resp))
}
