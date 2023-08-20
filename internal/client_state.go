package internal

import (
	"strconv"
	"strings"
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
	var builder strings.Builder

	builder.WriteString(CLIENT_REQUEST_PREFIX)
	builder.WriteString(DELIMETER)

	builder.WriteString(input)
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(state.currentRequestNumber))
	builder.WriteString(DELIMETER)
	state.currentRequestNumber += 1

	return builder.String()
}
