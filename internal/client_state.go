package internal

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
