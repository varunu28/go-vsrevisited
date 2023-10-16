package internal

import (
	"strconv"

	Text "github.com/linkdotnet/golang-stringbuilder"
)

// ClientState struct consists of the state that is maintained on the client side. It consists of:
// - configuration: Sorted array containing ports of all replicas
// - clientId: id associated with the client. In this implementation, we are using the port as clientId
// - currentViewNumber: used to track the primary replica
// - currentRequestNumber: A monotonically increasing integer that is associated with each client request
type ClientState struct {
	configuration        []int
	clientId             int
	currentViewNumber    int
	currentRequestNumber int
}

// NewClientState creates a new instance of ClientState on a particular port
func NewClientState(port int) *ClientState {
	var configuration [NUMBER_OF_NODES]int
	for i := 0; i < NUMBER_OF_NODES; i++ {
		configuration[i] = STARTING_PORT + i
	}
	return &ClientState{
		configuration:        configuration[:],
		clientId:             port,
		currentViewNumber:    0,
		currentRequestNumber: 0,
	}
}

// BuildClientRequest is a utility function that creates a client request in the format expected by replica node.
func (state *ClientState) BuildClientRequest(input string) string {
	sb := Text.StringBuilder{}

	sb.Append(CLIENT_REQUEST_PREFIX)
	sb.Append(DELIMETER)
	sb.Append(input)
	sb.Append(DELIMETER)
	sb.Append(strconv.Itoa(state.currentRequestNumber))

	state.currentRequestNumber += 1

	return sb.ToString()
}

// Broadcast sends a UDP message to all the replica nodes
func (state *ClientState) Broadcast(clientRequest string, udpHandler *UdpHandler) {
	for i := 0; i < NUMBER_OF_NODES; i++ {
		udpHandler.Send(clientRequest, STARTING_PORT+i)
	}
}
