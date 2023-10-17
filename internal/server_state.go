package internal

import (
	"strconv"
	"strings"

	Text "github.com/linkdotnet/golang-stringbuilder"
)

// ClientTableValue contains the client request, the number associated with the requested & response associated with the request if it is processed.
type ClientTableValue struct {
	Request       string
	RequestNumber int
	Response      string
}

type doViewChange struct {
	oldViewNumber   int
	newViewNumber   int
	operationNumber int
	commitNumber    int
	logs            []string
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
	viewChangeMap   map[int][]int
	doViewChangeMap map[int]doViewChange
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
		viewChangeMap:   map[int][]int{},
		doViewChangeMap: make(map[int]doViewChange),
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
	sb := Text.StringBuilder{}
	entry := sb.Append(command).
		Append(LOG_DELIMETER).
		AppendInt(requestNumber).
		Append(LOG_DELIMETER).
		AppendInt(port).
		ToString()
	state.log = append(state.log, entry)
	// Update client table
	ctValue := &ClientTableValue{
		Request:       command,
		RequestNumber: requestNumber,
		Response:      "",
	}
	state.clientTable[port] = *ctValue
}

func (state *ServerState) InitializeVoteTable(clientPort int) {
	state.voteTable[clientPort] = make(map[int]bool)
}

// Broadcast is invoked by the leader node to send a message to all peer nodes except itself.
func (state *ServerState) Broadcast(message string, udpHandler *UdpHandler) {
	for i := 0; i < NUMBER_OF_NODES; i++ {
		if i != state.replicaNumber {
			udpHandler.Send(message, STARTING_PORT+i)
		}
	}
}

func (state *ServerState) RecordViewChange(port int, viewNumber int) bool {
	state.viewChangeMap[viewNumber] = append(state.viewChangeMap[viewNumber], port)
	return len(state.viewChangeMap[viewNumber]) >= NUMBER_OF_NODES/2
}

// RecordPrepareResponse records the response from a replica node & returns a boolean value representing if quorum has been reached
func (state *ServerState) RecordPrepareResponse(port int, replicaId int) bool {
	state.voteTable[port][replicaId] = true
	return len(state.voteTable[port]) >= NUMBER_OF_NODES/2
}

func (state *ServerState) RecordCommit(port int, response string) {
	// increment commit number
	state.IncrementCommitNumber()
	// update client table
	ctValue := state.clientTable[port]
	ctValue.Response = response
	state.clientTable[port] = ctValue
}

func (state *ServerState) IncrementCommitNumber() {
	state.commitNumber += 1
}

func (state *ServerState) RecordDoViewChange(message string, port int) bool {
	parts := strings.Split(message, DELIMETER)

	oldViewNumber, _ := strconv.Atoi(parts[1])
	newViewNumber, _ := strconv.Atoi(parts[2])
	operationNumber, _ := strconv.Atoi(parts[3])
	commitNumber, _ := strconv.Atoi(parts[4])
	logs := strings.Split(parts[5], ",")
	doViewChange := &doViewChange{
		oldViewNumber:   oldViewNumber,
		newViewNumber:   newViewNumber,
		operationNumber: operationNumber,
		commitNumber:    commitNumber,
		logs:            logs,
	}
	state.doViewChangeMap[port] = *doViewChange
	return len(state.doViewChangeMap) >= NUMBER_OF_NODES/2
}

func (state *ServerState) UpdateForNewView() []string {
	// get nodes with highest old view number
	maxViewNum := 0
	for _, v := range state.doViewChangeMap {
		if maxViewNum < v.oldViewNumber {
			maxViewNum = v.oldViewNumber
		}
	}
	highestOldViewNodes := make([]int, 0)
	for k, v := range state.doViewChangeMap {
		if v.oldViewNumber == maxViewNum {
			highestOldViewNodes = append(highestOldViewNodes, k)
		}
	}
	// get node with highest operation number
	nodeWithHighestOpNumber := highestOldViewNodes[0]
	for _, node := range highestOldViewNodes {
		if state.doViewChangeMap[node].operationNumber > state.doViewChangeMap[nodeWithHighestOpNumber].operationNumber {
			nodeWithHighestOpNumber = node
		}
	}
	// list the logs that need to be committed by comparing the commit number
	state.log = state.doViewChangeMap[nodeWithHighestOpNumber].logs
	// change view number & operation number
	state.viewNumber = state.doViewChangeMap[nodeWithHighestOpNumber].newViewNumber
	state.operationNumber = state.doViewChangeMap[nodeWithHighestOpNumber].operationNumber
	// return logs to be committed
	latestCommitNumber := state.doViewChangeMap[nodeWithHighestOpNumber].commitNumber
	uncommittedLogs := state.log[state.operationNumber-1 : latestCommitNumber]
	// reset do view change map
	state.doViewChangeMap = make(map[int]doViewChange)

	return uncommittedLogs
}

func (state *ServerState) UpdateView(operationNumber int, viewNumber int, commitNumber int, logs []string) {
	state.viewNumber = viewNumber
	state.operationNumber = operationNumber
	state.commitNumber = commitNumber
	state.log = logs
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
		ToString()
}

// BuildPrepareRequest prepares a string represenatation of leader node's prepare request
func (state *ServerState) BuildPrepareRequest(command string, requestNumber int, port int) string {
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
		ToString()
}

// BuildCommitMessage prepares a string representation of leader node's commit message
func (state *ServerState) BuildCommitMessage(requestNumber int, port int) string {
	sb := Text.StringBuilder{}

	return sb.Append(COMMIT_MESSAGE_PREFIX).
		Append(DELIMETER).
		Append(strconv.Itoa(state.viewNumber)).
		Append(DELIMETER).
		Append(strconv.Itoa(requestNumber)).
		Append(DELIMETER).
		Append(strconv.Itoa(port)).
		ToString()
}

// BuildCatchupRequest prepares a string representation of replica node's catchup request
func (state *ServerState) BuildCatchupRequest(operationNumber int) string {
	sb := Text.StringBuilder{}

	return sb.Append(CATCHUP_REQUEST_PREFIX).
		Append(DELIMETER).
		AppendInt(state.operationNumber).
		Append(DELIMETER).
		AppendInt(operationNumber).
		ToString()
}

// BuildCatchupResponse prepares a string representation of catchup response
func (state *ServerState) BuildCatchupResponse(replicaOperationNumber int, laggingOperationNumber int) string {
	sb := Text.StringBuilder{}

	return sb.
		Append(CATCHUP_RESPONSE_PREFIX).
		Append(DELIMETER).
		AppendInt(state.commitNumber).
		Append(DELIMETER).
		Append(strings.Join(state.log[replicaOperationNumber:laggingOperationNumber-1], ",")).
		ToString()
}

func (state *ServerState) BuildStartViewChangeRequest() string {
	sb := Text.StringBuilder{}

	return sb.Append(START_VIEW_CHANGE_PREFIX).
		Append(DELIMETER).
		AppendInt(state.viewNumber).
		ToString()
}

func (state *ServerState) BuildDoViewChangeRequest(oldViewNumber int, newViewNumber int) string {
	sb := Text.StringBuilder{}

	return sb.Append(DO_VIEW_CHANGE_PREFIX).
		Append(DELIMETER).
		AppendInt(oldViewNumber).
		Append(DELIMETER).
		AppendInt(newViewNumber).
		Append(DELIMETER).
		AppendInt(state.operationNumber).
		Append(DELIMETER).
		AppendInt(state.commitNumber).
		Append(DELIMETER).
		Append(strings.Join(state.log, ",")).
		ToString()
}

func (state *ServerState) BuildStartViewRequest() string {
	sb := Text.StringBuilder{}

	return sb.Append(START_VIEW_PREFIX).
		Append(DELIMETER).
		AppendInt(state.operationNumber).
		Append(DELIMETER).
		AppendInt(state.viewNumber).
		Append(DELIMETER).
		AppendInt(state.commitNumber).
		Append(DELIMETER).
		Append(strings.Join(state.log, ",")).
		ToString()
}

func (state *ServerState) UpdateStatus(status string) {
	state.status = status
}

func (state *ServerState) GetStatus() string {
	return state.status
}
