package internal

import (
	"fmt"
	"strconv"
	"strings"
)

// VsServer is a struct used to communicate with client & peer nodes.
// It is also responsible for maintaining the state associated with a replica
type VsServer struct {
	udpHandler *UdpHandler
	state      *ServerState
}

// NewVsServer creates an instance of VsServer on a given port.
// It returns an error if the creation process fails.
func NewVsServer(port int) (*VsServer, error) {
	udpHandler, err := NewUdpHandler(port)
	if err != nil {
		return nil, err
	}
	return &VsServer{
		udpHandler: udpHandler,
		state:      NewServerState(port),
	}, nil
}

// Start runs an infinite loop where it listens on its port & then processes any messages that it receives.
func (server *VsServer) Start() {
	for {
		message, err := server.udpHandler.Receive()
		if err != nil {
			panic(err)
		}
		go server.handleMessage(message)
	}
}

func (server *VsServer) handleMessage(message UdpMessage) {
	fmt.Println("received message: ", message)
	parts := strings.Split(message.Message, DELIMETER)
	msgType := parts[0]
	if msgType == CLIENT_REQUEST_PREFIX {
		if server.state.viewNumber%NUMBER_OF_NODES != server.state.replicaNumber {
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
	if viewNumber != server.state.viewNumber {
		return
	}
	if operationNumber != server.state.operationNumber+1 {
		return
	}
	// Update client state
	server.state.RecordRequest(command, requestNumber, port)
	// Send a vote acknowledging the request processing
	server.udpHandler.Send(server.state.BuildPrepareResponse(operationNumber, port), fromPort)
}

func (server *VsServer) handlePrepareResponse(viewNumber int, operationNumber int, port int, replicaId int) {
	quorum := server.state.RecordPrepareResponse(port, replicaId)
	if quorum {
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
		server.udpHandler.Send(SERVER_RESPONSE_PREFIX+DELIMETER+response, port)

		// Broadcast about commit
		commitMessage := server.state.BuildCommitMessage()
		server.state.Broadcast(commitMessage, server.udpHandler)
	}
}

func (server *VsServer) handleCommitMessage(viewNumber int, requestNumber int, port int) {
	if viewNumber != server.state.viewNumber {
		return
	}
	clientTableValue, exists := server.state.GetClientTableValue(port)
	if exists {
		if clientTableValue.RequestNumber != requestNumber {
			fmt.Printf("got out of range request number. got %d current %d\n", requestNumber, clientTableValue.RequestNumber)
			return
		}
		// perform commit
		response := server.performServerOperation(clientTableValue.Request)
		server.state.RecordCommit(port, response)
	}
}

func (server *VsServer) performServerOperation(request string) string {
	response := "sample_response"
	fmt.Println("Performing request: " + request + " RESPONSE: " + response)
	return response
}
