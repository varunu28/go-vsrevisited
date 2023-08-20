package internal

import (
	"strconv"

	Text "github.com/linkdotnet/golang-stringbuilder"
)

type ClientState struct {
	configuration        []int
	clientId             int
	currentViewNumber    int
	currentRequestNumber int
}

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

func (state *ClientState) BuildClientRequest(input string) string {
	sb := Text.StringBuilder{}

	sb.Append(CLIENT_REQUEST_PREFIX)
	sb.Append(DELIMETER)
	sb.Append(input)
	sb.Append(DELIMETER)
	sb.Append(strconv.Itoa(state.currentRequestNumber))
	sb.Append(DELIMETER)

	state.currentRequestNumber += 1

	return sb.ToString()
}
