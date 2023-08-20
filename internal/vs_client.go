package internal

import (
	"bufio"
	"fmt"
	"os"
)

type VsClient struct {
	udp_handler *UdpHandler
	reader      *bufio.Reader
	state       *ClientState
}

func NewVsClient(port int) (*VsClient, error) {
	udp_handler, err := NewUdpHandler(port)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(os.Stdin)
	return &VsClient{
		udp_handler: udp_handler,
		reader:      reader,
		state:       NewClientState(port),
	}, nil
}

func (client *VsClient) Start() {
	for {
		input, err := client.reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		err = client.udp_handler.Send(input, 8000)
		if err != nil {
			fmt.Println(err.Error())
		}
		message, err := client.udp_handler.Receive()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("response: " + message.Message)
	}
}
