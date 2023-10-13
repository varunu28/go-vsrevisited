package internal

import (
	"strconv"

	Text "github.com/linkdotnet/golang-stringbuilder"
)

const (
	NORMAL      = "normal"
	VIEW_CHANGE = "view change"
	RECOVERING  = "recovering"
)

// ClientTableValue contains the client request, the number associated with the requested & response associated with the request if it is processed.
type ClientTableValue struct {
	Request       string
	RequestNumber int
	Response      string
}

// ServerState struct is tied to each replica. It consists of:
// - configuration: Sorted array containing ports of all replicas
// - viewNumber: current view number
// - status: current status associated with the replica
// - operationNumber: monotonically increasing counter associated to each request
// - log: an array containing all requests. Size of log is same as that of operationNumber
// - commitNumber: operationNumber associated with most recently committed operation
// - clientTable: A hashmap to record the latest ClientTableValue for each client
// - replicaNumber: index of replica in the configuration
// - voteTable: A hashmap for recording votes for each client request. This is used to establish quorum for a client request
type ServerState struct {
	configuration   []int
	viewNumber      int
	status          string
	operationNumber int
	log             []string
	commitNumber    int
	clientTable     map[int]ClientTableValue
	replicaNumber   int
	voteTable       map[int]map[int]bool
}

// NewServerState creates a new instance of ServerState on a given port number
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
		voteTable:       make(map[int]map[int]bool),
	}
}

// GetClientTableValue retrieves ClientTableValue for a client
func (state *ServerState) GetClientTableValue(port int) (ClientTableValue, bool) {
	val, exists := state.clientTable[port]
	return val, exists
}

// RecordRequest updates the server state for a new client request.
// It is invoked either by leader replica for processing new client request or by replica nodes while processing PrepareRequest from leader node
func (state *ServerState) RecordRequest(command string, requestNumber int, port int) {
	// Increment operation number
	state.operationNumber += 1
	// Add request to log
	state.log = append(state.log, command)
	// Update client table
	ctValue := &ClientTableValue{
		Request:       command,
		RequestNumber: requestNumber,
		Response:      "",
	}
	state.clientTable[port] = *ctValue
}

// Broadcast is invoked by the leader node to send a message to all peer nodes.
func (state *ServerState) Broadcast(command string, requestNumber int, port int, udpHandler *UdpHandler) {
	state.voteTable[port] = make(map[int]bool)
	prepareRequest := state.buildPrepareRequest(command, requestNumber, port)
	for i := 0; i < NUMBER_OF_NODES; i++ {
		if i != state.replicaNumber {
			udpHandler.Send(prepareRequest, STARTING_PORT+i)
		}
	}
}

// RecordPrepareResponse records the response from a replica node & returns a boolean value representing if quorum has been reached
func (state *ServerState) RecordPrepareResponse(port int, replicaId int) bool {
	state.voteTable[port][replicaId] = true
	return len(state.voteTable[port]) >= NUMBER_OF_NODES/2
}

func (state *ServerState) RecordCommit(port int, response string) {
	// increment commit number
	state.commitNumber += 1
	// update client table
	ctValue := state.clientTable[port]
	ctValue.Response = response
	state.clientTable[port] = ctValue
}

// BuildPrepareResponse prepares a string representation for response of PrepareRequest
func (state *ServerState) BuildPrepareResponse(operationNumber int, port int) string {
	sb := Text.StringBuilder{}

	return sb.Append(PREPARE_RESPONSE_PREFIX).
		Append(DELIMETER).
		Append(strconv.Itoa(state.viewNumber)).
		Append(DELIMETER).
		Append(strconv.Itoa(operationNumber)).
		Append(DELIMETER).
		Append(strconv.Itoa(port)).
		Append(DELIMETER).
		Append(strconv.Itoa(state.configuration[state.replicaNumber])).
		Append(DELIMETER).
		ToString()
}

func (state *ServerState) buildPrepareRequest(command string, requestNumber int, port int) string {
	sb := Text.StringBuilder{}

	return sb.Append(PREPARE_REQUEST_PREFIX).
		Append(DELIMETER).
		Append(strconv.Itoa(state.viewNumber)).
		Append(DELIMETER).
		Append(command).
		Append(DELIMETER).
		Append(strconv.Itoa(requestNumber)).
		Append(DELIMETER).
		Append(strconv.Itoa(port)).
		Append(DELIMETER).
		Append(strconv.Itoa(state.operationNumber)).
		Append(DELIMETER).
		Append(strconv.Itoa(state.commitNumber)).
		Append(DELIMETER).
		ToString()
}
