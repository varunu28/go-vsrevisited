package internal

import "fmt"

type VsServer struct {
	udpHandler *UdpHandler
	state      *ServerState
}

func NewVsServer(port int) (*VsServer, error) {
	udpHandler, err := NewUdpHandler(port)
	if err != nil {
		return nil, err
	}
	return &VsServer{
		udpHandler: udpHandler,
		state:      NewServerState(port),
	}, nil
}

func (server *VsServer) Start() {
	for {
		message, err := server.udpHandler.Receive()
		if err != nil {
			panic(err)
		}
		go server.handleMessage(message)
	}
}

func (server *VsServer) handleMessage(message UdpMessage) {
	fmt.Println("received message: ", message)
	response := "Received: " + message.Message
	server.udpHandler.Send(response, message.FromPort)
}
