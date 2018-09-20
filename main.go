package main

import (
	"flag"

	"github.com/adiclepcea/trivia/server"
)

func main() {

	port := flag.String("port", ":8080", "the port to use in your server")

	flag.Parse()

	server.Serve(*port)
}
