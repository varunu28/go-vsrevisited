package internal

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// VsClient is a wraper struct that is responsible for:
// - Reading input from user through System input
// - Sending message through UDP protocol
// - Maintaining ClientState
type VsClient struct {
	udp_handler *UdpHandler
	reader      *bufio.Reader
	state       *ClientState
}

// NewVsClient creates an instance of VsClient on a specified port.
// It returns an error if the creation fails.
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

// Start runs an infinite loop which performs following steps in order:
// - Reads input from user
// - Sends a message to leader node through UDP
// - Receives the response from leader
// - Prints the response
func (client *VsClient) Start() {
	for {
		// read user input
		input, err := client.reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		// send message to leader node
		clientRequest := client.state.BuildClientRequest(input)
		err = client.udp_handler.Send(clientRequest, client.state.GetLeaderPort())
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// read response for message
		client.receive(clientRequest)
	}
}

func (client *VsClient) receive(clientRequest string) {
	message, err := client.udp_handler.RecieveWithTimeout(1 * time.Second)
	if err != nil {
		// if timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// broadcast to all nodes & receive
			client.state.Broadcast(clientRequest, client.udp_handler)
			// invoke receive
			client.receive(clientRequest)
		} else {
			fmt.Println("error: ", err)
		}
	} else {
		parts := strings.Split(message.Message, DELIMETER)
		viewNumber, _ := strconv.Atoi(parts[1])
		response := parts[2]
		client.state.RecordViewNumber(viewNumber)
		fmt.Println("response: " + response)
	}
}
