package main

import (
	"os"
	"strconv"
	"vsrevisited/internal"
)

func main() {
	if len(os.Args) != 3 {
		panic("need two arguments. Type(client/server) & a port")
	}
	t := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic("port should be an integer")
	}
	if t == "client" {
		client, err := internal.NewVsClient(port)
		if err != nil {
			panic("error while creating new client: " + err.Error())
		}
		client.Start()
	} else if t == "server" {
		server, err := internal.NewVsServer(port)
		if err != nil {
			panic("error while creating new server" + err.Error())
		}
		server.Start()
	} else {
		panic("invalid type for runner")
	}
}
