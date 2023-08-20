package internal

import (
	"fmt"
	"strconv"
	"strings"
)

type VsServer struct {
	udpHandler *UdpHandler
	state      *ServerState
}

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
	}
}

func (server *VsServer) handleClientRequest(command string, currentRequestNumber string, port int) {
	// validate request
	reqNo, err := strconv.Atoi(currentRequestNumber)
	if err != nil {
		server.udpHandler.Send(SERVER_RESPONSE_PREFIX+DELIMETER+SERVER_RESPONSE_NON_NUMERIC_REQUEST_NUMBER, port)
		return
	}
	if !server.state.IsClientRequestValid(port, reqNo) {
		server.udpHandler.Send(SERVER_RESPONSE_PREFIX+DELIMETER+SERVER_RESPONSE_INVALID_REQUEST_NUMER, port)
		return
	}
	// check if client table contains the request with requestId
	existingResponse, exists := server.state.GetClientResponseByRequestNumber(port, reqNo)
	if exists {
		if existingResponse != "" {
			server.udpHandler.Send(SERVER_RESPONSE_PREFIX+DELIMETER+existingResponse, port)
		}
		return
	}
	// Update client state
	server.state.RecordRequest(command, reqNo, port)
	// Broadcast for vote
	server.state.Broadcast(command, reqNo, port, server.udpHandler)
}

func (server *VsServer) handlePrepareRequest(viewNumber int, command string, requestNumber int, port int, operationNumber int, commitNumber int, fromPort int) {
	if viewNumber != server.state.viewNumber {
		return
	}
	if operationNumber != server.state.operationNumber+1 {
		return
	}
	server.state.RecordRequest(command, requestNumber, port)
	server.udpHandler.Send(server.state.BuildPrepareResponse(operationNumber, port), fromPort)
}

func (server *VsServer) handlePrepareResponse(viewNumber int, operationNumber int, port int, replicaId int) {
	quorum := server.state.RecordPrepareResponse(port, replicaId)
	if quorum {
		// perform the operation
		request := server.state.GetClientRequest(port)
		fmt.Println("Performing request: " + request)
		response := "sample_response"

		server.state.RecordCommit(port, response)

		// send response to client
		server.udpHandler.Send(SERVER_RESPONSE_PREFIX+DELIMETER+response, port)
	}
}
