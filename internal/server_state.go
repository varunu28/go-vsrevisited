package internal

const (
	NORMAL      = "normal"
	VIEW_CHANGE = "view change"
	RECOVERING  = "recovering"
)

type ClientTableValue struct {
	Request  string
	Response string
}

type ServerState struct {
	configuration   []int
	viewNumber      int
	status          string
	operationNumber int
	log             []string
	commitNumber    int
	clientTable     map[int]ClientTableValue
	replicaNumber   int
}

func NewServerState(port int) *ServerState {
	var configuration [NUMBER_OF_NODES]int
	var replicaNumber = 0
	for i := 0; i < NUMBER_OF_NODES; i++ {
		configuration[i] = STARTING_PORT + i
		if configuration[i] == port {
			replicaNumber = i
		}
	}
	return &ServerState{
		configuration:   configuration[:],
		viewNumber:      0,
		status:          NORMAL,
		operationNumber: 0,
		log:             make([]string, 0),
		commitNumber:    0,
		clientTable:     make(map[int]ClientTableValue),
		replicaNumber:   replicaNumber,
	}
}
