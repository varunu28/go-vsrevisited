package internal

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type bufferedRequest struct {
	command       string
	requestNumber int
	clientPort    int
	commitNumber  int
	serverPort    int
}

// VsServer is a struct used to communicate with client & peer nodes.
// It is also responsible for maintaining the state associated with a replica
type VsServer struct {
	udpHandler    *UdpHandler
	state         *ServerState
	database      *Database
	serverTimeout *ServerTimeout
	requestBuffer []bufferedRequest
	mu            sync.Mutex
}

// NewVsServer creates an instance of VsServer on a given port.
// It returns an error if the creation process fails.
func NewVsServer(port int) (*VsServer, error) {
	udpHandler, err := NewUdpHandler(port)
	if err != nil {
		return nil, err
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	timeoutInterval := rand.Intn(int(MAX_TIMEOUT)-int(MIN_TIMEOUT)) + int(MIN_TIMEOUT)
	serverTimeout := NewServerTimeout(timeoutInterval)

	return &VsServer{
		udpHandler:    udpHandler,
		state:         NewServerState(port),
		database:      NewDatabase(),
		serverTimeout: serverTimeout,
		requestBuffer: make([]bufferedRequest, 0),
		mu:            sync.Mutex{},
	}, nil
}

// Start runs an infinite loop where it listens on its port & then processes any messages that it receives.
func (server *VsServer) Start() {
	go server.serverTimer()
	for {
		message, err := server.udpHandler.Receive()
		if err != nil {
			panic(err)
		}
		go server.handleMessage(message)
	}
}

func (server *VsServer) handleMessage(message UdpMessage) {
	fmt.Println("[received message] ", message.Message)
	parts := strings.Split(message.Message, DELIMETER)
	msgType := parts[0]
	if msgType == CLIENT_REQUEST_PREFIX {
		if !server.isLeader() {
			return
		}
		server.handleClientRequest(parts[1], parts[2], message.FromPort)
	} else if msgType == PREPARE_REQUEST_PREFIX {
		viewNumber, _ := strconv.Atoi(parts[1])
		requestNumber, _ := strconv.Atoi(parts[3])
		port, _ := strconv.Atoi(parts[4])
		operationNumber, _ := strconv.Atoi(parts[5])
		commitNumber, _ := strconv.Atoi(parts[6])
		server.handlePrepareRequest(viewNumber, parts[2], requestNumber, port, operationNumber, commitNumber, message.FromPort)
	} else if msgType == PREPARE_RESPONSE_PREFIX {
		viewNumber, _ := strconv.Atoi(parts[1])
		operationNumber, _ := strconv.Atoi(parts[2])
		port, _ := strconv.Atoi(parts[3])
		replicaId, _ := strconv.Atoi(parts[4])
		server.handlePrepareResponse(viewNumber, operationNumber, port, replicaId)
	} else if msgType == COMMIT_MESSAGE_PREFIX {
		viewNumber, _ := strconv.Atoi(parts[1])
		requestNumber, _ := strconv.Atoi(parts[2])
		port, _ := strconv.Atoi(parts[3])
		server.handleCommitMessage(viewNumber, requestNumber, port)
	} else if msgType == CATCHUP_REQUEST_PREFIX {
		repOpNo, _ := strconv.Atoi(parts[1])
		lagOpNo, _ := strconv.Atoi(parts[2])
		server.handleCatchupMessage(repOpNo, lagOpNo, message.FromPort)
	} else if msgType == CATCHUP_RESPONSE_PREFIX {
		commitNo, _ := strconv.Atoi(parts[1])
		backupLogs := strings.Split(parts[2], ",")
		server.processBackupLogs(backupLogs, commitNo)
	} else if msgType == START_VIEW_CHANGE_PREFIX {
		updatedViewNumber, _ := strconv.Atoi(parts[1])
		port := message.FromPort
		server.processStartViewChangeMessage(updatedViewNumber, port)
	} else if msgType == DO_VIEW_CHANGE_PREFIX {
		server.processDoViewChangeMessage(message.Message, message.FromPort)
	} else if msgType == START_VIEW_PREFIX {
		operationNumber, _ := strconv.Atoi(parts[1])
		viewNumber, _ := strconv.Atoi(parts[2])
		commitNumber, _ := strconv.Atoi(parts[3])
		logs := strings.Split(parts[4], ",")
		server.startNewView(operationNumber, viewNumber, commitNumber, logs)
	}
}

func (server *VsServer) handleClientRequest(command string, currentRequestNumber string, port int) {
	// validate request
	reqNo, err := strconv.Atoi(currentRequestNumber)
	if err != nil {
		server.udpHandler.Send(SERVER_RESPONSE_PREFIX+DELIMETER+SERVER_RESPONSE_NON_NUMERIC_REQUEST_NUMBER, port)
		return
	}
	// check the state of existing request in ClientTable for client
	clientTableValue, exists := server.state.GetClientTableValue(port)
	if exists {
		// error for sending an already processed request number
		if clientTableValue.RequestNumber > reqNo {
			server.udpHandler.Send(SERVER_RESPONSE_PREFIX+DELIMETER+SERVER_RESPONSE_INVALID_REQUEST_NUMER, port)
			return
		}
		if clientTableValue.RequestNumber == reqNo {
			// send the processed response to client for the processed request
			if clientTableValue.Response != "" {
				server.udpHandler.Send(SERVER_RESPONSE_PREFIX+DELIMETER+clientTableValue.Response, port)
			}
			return
		}
	}
	// Update client state
	server.state.RecordRequest(command, reqNo, port)
	server.state.InitializeVoteTable(port)

	// Broadcast for vote
	prepareRequest := server.state.BuildPrepareRequest(command, reqNo, port)
	server.state.Broadcast(prepareRequest, server.udpHandler)
}

func (server *VsServer) handlePrepareRequest(viewNumber int, command string, requestNumber int, port int, operationNumber int, commitNumber int, fromPort int) {
	if viewNumber < server.state.viewNumber {
		return
	}
	// view change has occurred
	if viewNumber > server.state.viewNumber {
		server.state.viewNumber = viewNumber
	}
	// reset timeout as we received a ping from leader replica
	server.serverTimeout.Reset <- struct{}{}
	// if replica is in recovery state then add the request to buffer
	if server.state.GetStatus() == RECOVERING {
		buffReq := &bufferedRequest{
			command:       command,
			requestNumber: requestNumber,
			clientPort:    port,
			commitNumber:  commitNumber,
			serverPort:    fromPort,
		}
		server.requestBuffer = append(server.requestBuffer, *buffReq)
		return
	}

	if operationNumber == server.state.operationNumber+1 {
		// Update client state
		server.state.RecordRequest(command, requestNumber, port)
		// Send a vote acknowledging the request processing
		server.udpHandler.Send(server.state.BuildPrepareResponse(operationNumber, port), fromPort)
	} else if operationNumber > server.state.operationNumber+1 {
		// update state to catching up
		server.state.UpdateStatus(RECOVERING)

		// push request to request_buffer
		buffReq := &bufferedRequest{
			command:       command,
			requestNumber: requestNumber,
			clientPort:    port,
			commitNumber:  commitNumber,
			serverPort:    fromPort,
		}
		server.requestBuffer = append(server.requestBuffer, *buffReq)

		// send catch up request to leader
		server.udpHandler.Send(server.state.BuildCatchupRequest(operationNumber), fromPort)
	}
}

func (server *VsServer) handlePrepareResponse(viewNumber int, operationNumber int, port int, replicaId int) {
	quorum := server.state.RecordPrepareResponse(port, replicaId)
	if quorum {
		// locking is required as we can get concurrent prepare response and we want to perform the commit & broadcast about it at most once
		// Hence while checking the existing response & updating the response, locking allows only one request to go through the commit & broadcast phase
		// When the next thread acquires the thread post commit, it will see the existing response as non-empty & return instead of performing duplicate commit & broadcast
		server.mu.Lock()
		defer server.mu.Unlock()

		// perform the operation
		clientTableValue, _ := server.state.GetClientTableValue(port)
		// Don't process request if it is already processed. These are lagging nodes which are late to respond to PrepareRequest
		if clientTableValue.Response != "" {
			return
		}

		// perform commit
		response := server.performServerOperation(clientTableValue.Request)
		server.state.RecordCommit(port, response)

		// send response to client
		server.udpHandler.Send(server.state.BuildClientResponse(response), port)

		// Broadcast about commit
		commitMessage := server.state.BuildCommitMessage(clientTableValue.RequestNumber, port)
		server.state.Broadcast(commitMessage, server.udpHandler)
	}
}

func (server *VsServer) handleCommitMessage(viewNumber int, requestNumber int, port int) {
	if viewNumber != server.state.viewNumber {
		return
	}
	// reset timeout as we received a ping from leader replica
	server.serverTimeout.Reset <- struct{}{}

	if server.state.GetStatus() != NORMAL {
		return
	}

	clientTableValue, exists := server.state.GetClientTableValue(port)
	if exists {
		if clientTableValue.RequestNumber != requestNumber {
			fmt.Printf("[commit_message_error] got out of range request number. got %d current %d\n", requestNumber, clientTableValue.RequestNumber)
			return
		}
		// perform commit
		response := server.performServerOperation(clientTableValue.Request)
		server.state.RecordCommit(port, response)
	}
}

func (server *VsServer) handleCatchupMessage(replicaOperationNumber int, laggingOperationNumber int, fromPort int) {
	server.udpHandler.Send(server.state.BuildCatchupResponse(replicaOperationNumber, laggingOperationNumber), fromPort)
}

func (server *VsServer) processBackupLogs(logs []string, commitNumber int) {
	for _, log := range logs {
		splits := strings.Split(log, LOG_DELIMETER)
		reqNo, _ := strconv.Atoi(splits[1])
		port, _ := strconv.Atoi(splits[2])
		server.state.RecordRequest(splits[0], reqNo, port)
	}
	// process backed up requests
	for _, entry := range server.requestBuffer {
		server.state.RecordRequest(entry.command, entry.requestNumber, entry.clientPort)
	}
	// clear the buffer
	server.requestBuffer = []bufferedRequest{}
	// get updated on latest commit
	for i := 0; i < commitNumber; i++ {
		log := server.state.log[i]
		server.commitLog(log)
	}
	// update status
	server.state.UpdateStatus(NORMAL)
}

func (server *VsServer) commitLog(log string) {
	splits := strings.Split(log, LOG_DELIMETER)
	command := splits[0]
	reqNo, _ := strconv.Atoi(splits[1])
	port, _ := strconv.Atoi(splits[2])
	response := server.performServerOperation(command)
	ctValue, _ := server.state.GetClientTableValue(port)
	if ctValue.RequestNumber == reqNo {
		server.state.RecordCommit(port, response)
	} else {
		server.state.IncrementCommitNumber()
	}
}

func (server *VsServer) performServerOperation(request string) string {
	response := server.database.PerformOperation(request)
	return response
}

func (server *VsServer) processStartViewChangeMessage(updatedViewNumber int, fromPort int) {
	if updatedViewNumber >= server.state.viewNumber {
		if updatedViewNumber > server.state.viewNumber {
			server.state.viewNumber = updatedViewNumber
			server.state.UpdateStatus(VIEW_CHANGE)
			startViewChangeReq := server.state.BuildStartViewChangeRequest()
			server.state.Broadcast(startViewChangeReq, server.udpHandler)
		}
		majority := server.state.RecordViewChange(fromPort, updatedViewNumber)
		if majority {
			server.initiateDoViewChange(updatedViewNumber)
		}
	}
}

func (server *VsServer) initiateDoViewChange(viewNumber int) {
	newLeaderPort := STARTING_PORT + viewNumber%NUMBER_OF_NODES
	// Assuming that the next replica in order is alive, old view number
	// will be one less than the updated view number.
	// This is not necessarily true as the next replica in order can also fail in which case,
	// the replica next to it will be elected as leader
	oldViewNumber := viewNumber - 1
	doViewChangeRequest := server.state.BuildDoViewChangeRequest(oldViewNumber, viewNumber)
	server.udpHandler.Send(doViewChangeRequest, newLeaderPort)
}

func (server *VsServer) processDoViewChangeMessage(message string, port int) {
	server.mu.Lock()
	defer server.mu.Unlock()

	if server.state.GetStatus() == NORMAL {
		return
	}

	majority := server.state.RecordDoViewChange(message, port)
	if majority {
		uncommitedLogs := server.state.UpdateForNewView()
		// commit any pending logs
		for _, log := range uncommitedLogs {
			server.commitLog(log)
		}
		// update status to normal
		server.state.UpdateStatus(NORMAL)
		// broadcast start view message
		startViewRequest := server.state.BuildStartViewRequest()
		server.state.Broadcast(startViewRequest, server.udpHandler)
	}
}

func (server *VsServer) startNewView(operationNumber int, viewNumber int, commitNumber int, logs []string) {
	server.state.UpdateView(operationNumber, viewNumber, commitNumber, logs)
	server.state.UpdateStatus(NORMAL)
	server.serverTimeout.Reset <- struct{}{}
}

func (server *VsServer) serverTimer() {
	for {
		select {
		case <-server.serverTimeout.Timeout.C:
			if server.isLeader() {
				server.serverTimeout.Reset <- struct{}{}
			} else {
				// perform view change
				fmt.Println("[replica_error] leader server timed out")
				if server.state.GetStatus() == NORMAL {
					server.startViewChange()
				}
			}
		case <-server.serverTimeout.Reset:
			server.serverTimeout.Timeout.Reset(time.Duration(server.serverTimeout.TimeoutInterval) * time.Millisecond)
		}
	}
}

func (server *VsServer) startViewChange() {
	// update state for view change
	server.state.viewNumber += 1
	server.state.UpdateStatus(VIEW_CHANGE)
	// Broadcast start view change request
	startViewChangeReq := server.state.BuildStartViewChangeRequest()
	server.state.Broadcast(startViewChangeReq, server.udpHandler)
}

func (server *VsServer) isLeader() bool {
	return server.state.viewNumber%NUMBER_OF_NODES == server.state.replicaNumber
}
