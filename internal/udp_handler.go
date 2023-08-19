package internal

import (
	"net"
	"strconv"
)

type UdpHandler struct {
	socket *net.UDPConn
}

type UdpMessage struct {
	Message  string
	FromPort int
}

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

func (u *UdpHandler) Receive() (UdpMessage, error) {
	data := make([]byte, 1024)
	_, addr, err := u.socket.ReadFromUDP(data)
	if err != nil {
		return UdpMessage{}, err
	}

	message := string(data)
	return UdpMessage{
		Message:  message,
		FromPort: addr.Port,
	}, nil
}

func (u *UdpHandler) Send(message string, port int) error {
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	_, err = u.socket.WriteToUDP([]byte(message), clientAddr)
	return err
}

func (u *UdpHandler) Close() error {
	return u.socket.Close()
}
