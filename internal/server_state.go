package internal

import (
	"strconv"
	"strings"
)

const (
	NORMAL      = "normal"
	VIEW_CHANGE = "view change"
	RECOVERING  = "recovering"
)

type ClientTableValue struct {
	Request       string
	RequestNumber int
	Response      string
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
	voteTable       map[int]map[int]bool
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
		voteTable:       make(map[int]map[int]bool),
	}
}

func (state *ServerState) IsClientRequestValid(port int, requestNumber int) bool {
	val, exists := state.clientTable[port]
	if !exists {
		return true
	}
	if val.RequestNumber <= requestNumber {
		return true
	}
	return false
}

func (state *ServerState) GetClientResponseByRequestNumber(port int, requestNumber int) (string, bool) {
	val, exists := state.clientTable[port]
	if !exists {
		return "", false
	}
	if val.RequestNumber < requestNumber {
		return "", false
	}
	return val.Response, true
}

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

func (state *ServerState) Broadcast(command string, requestNumber int, port int, udpHandler *UdpHandler) {
	state.voteTable[port] = make(map[int]bool)
	prepareRequest := state.BuildPrepareRequest(command, requestNumber, port)
	for i := 0; i < NUMBER_OF_NODES; i++ {
		if i != state.replicaNumber {
			udpHandler.Send(prepareRequest, STARTING_PORT+i)
		}
	}
}

func (state *ServerState) RecordPrepareResponse(port int, replicaId int) bool {
	state.voteTable[port][replicaId] = true
	return len(state.voteTable[port]) >= NUMBER_OF_NODES/2
}

func (state *ServerState) GetClientRequest(port int) string {
	return state.clientTable[port].Request
}

func (state *ServerState) RecordCommit(port int, response string) {
	// increment commit number
	state.commitNumber += 1
	// update client table
	ctValue := state.clientTable[port]
	ctValue.Response = response
}

func (state *ServerState) BuildPrepareRequest(command string, requestNumber int, port int) string {
	var builder strings.Builder

	builder.WriteString(PREPARE_REQUEST_PREFIX)
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(state.viewNumber))
	builder.WriteString(DELIMETER)

	builder.WriteString(command)
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(requestNumber))
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(port))
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(state.operationNumber))
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(state.commitNumber))
	builder.WriteString(DELIMETER)

	return builder.String()
}

func (state *ServerState) BuildPrepareResponse(operationNumber int, port int) string {
	var builder strings.Builder

	builder.WriteString(PREPARE_RESPONSE_PREFIX)
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(state.viewNumber))
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(operationNumber))
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(port))
	builder.WriteString(DELIMETER)

	builder.WriteString(strconv.Itoa(state.configuration[state.replicaNumber]))
	builder.WriteString(DELIMETER)

	return builder.String()
}
