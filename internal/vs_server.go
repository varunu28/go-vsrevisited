package internal

import "fmt"

type VrServer struct {
	udp_handler *UdpHandler
}

func NewVrServer(port int) (*VrServer, error) {
	udp_handler, err := NewUdpHandler(port)
	if err != nil {
		return nil, err
	}
	return &VrServer{
		udp_handler: udp_handler,
	}, nil
}

func (server *VrServer) Start() {
	for {
		message, err := server.udp_handler.Receive()
		if err != nil {
			panic(err)
		}
		fmt.Println("received message: ", message)
		response := "Received: " + message.Message
		server.udp_handler.Send(response, message.FromPort)
	}
}
