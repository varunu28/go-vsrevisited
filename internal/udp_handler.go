package internal

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"
)

// UdpHandler is a wrapper on top of udp client used to send & receive messages
type UdpHandler struct {
	socket *net.UDPConn
}

// UdpMessage is a record that describes the contents of a message & port from which the message is sent
type UdpMessage struct {
	Message  string
	FromPort int
}

// NewUpdHandler creates an instance of UdpHandler at a specific port.
// It takes a port as integer for input & returns an instance of UdpHandler.
// If there is an error while construction then it returns nil for instance & the error.
func NewUdpHandler(port int) (*UdpHandler, error) {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &UdpHandler{
		socket: conn,
	}, nil
}

// RecieveWithTimeout listens on the port for UdpHandler for a defined time duration
// If it receives a message within the time duration then it returns a UdpMessage instance
// by parsing the incoming message. Else it returns an error
func (u *UdpHandler) RecieveWithTimeout(timeout time.Duration) (UdpMessage, error) {
	err := u.socket.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		fmt.Println("error setting deadline: ", err)
	}
	data := make([]byte, 1024)
	_, addr, err := u.socket.ReadFromUDP(data)
	if err != nil {
		return UdpMessage{}, err
	}

	message := string(bytes.Trim(data, "\x00"))
	return UdpMessage{
		Message:  message,
		FromPort: addr.Port,
	}, nil
}

// Receive listens on the port for UdpHandler & returns a UdpMessage instance by parsing the incoming message
func (u *UdpHandler) Receive() (UdpMessage, error) {
	data := make([]byte, 1024)
	_, addr, err := u.socket.ReadFromUDP(data)
	if err != nil {
		return UdpMessage{}, err
	}

	message := string(bytes.Trim(data, "\x00"))
	return UdpMessage{
		Message:  message,
		FromPort: addr.Port,
	}, nil
}

// Send sends a message to a specified port. It returns an error if there is an error while sending the message.
func (u *UdpHandler) Send(message string, port int) error {
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	_, err = u.socket.WriteToUDP([]byte(message), clientAddr)
	return err
}

// Close closes the UDP socket associated with UdpHandler instance
func (u *UdpHandler) Close() error {
	return u.socket.Close()
}
