package tinynet

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/alphahorizonio/unisockets/pkg/unisockets"
)

type IP []byte

type TCPAddr struct {
	stringAddr string

	IP   IP
	Port int
	Zone string
}

func (t *TCPAddr) Network() string {
	return "tcp"
}

func (t *TCPAddr) String() string {
	return t.stringAddr
}

func ResolveTCPAddr(network, address string) (*TCPAddr, error) {
	parts := strings.Split(address, ":")

	ip := make([]byte, 4) // xxx.xxx.xxx.xxx
	for i, part := range strings.Split(parts[0], ".") {
		innerPart, err := strconv.Atoi(part)
		if err != nil {
			return nil, errors.New("could not parse IP")
		}

		ip[i] = byte(innerPart)
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.New("could not parse port")
	}

	return &TCPAddr{
		stringAddr: address,

		IP:   ip,
		Port: port,
		Zone: "",
	}, nil
}

func Listen(network, address string) (net.Listener, error) {
	laddr, err := ResolveTCPAddr(network, address)
	if err != nil {
		return TCPListener{}, err
	}

	return ListenTCP(network, laddr)
}

func ListenTCP(network string, laddr *TCPAddr) (*TCPListener, error) {
	// Create address
	serverAddress := unisockets.SockaddrIn{
		SinFamily: unisockets.PF_INET,
		SinPort:   unisockets.Htons(uint16(laddr.Port)),
		SinAddr: struct{ SAddr uint32 }{
			SAddr: binary.LittleEndian.Uint32(laddr.IP),
		},
	}

	// Create socket
	serverSocket, err := unisockets.Socket(unisockets.PF_INET, unisockets.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	// Bind
	if err := unisockets.Bind(serverSocket, &serverAddress); err != nil {
		return nil, err
	}

	// Listen
	if err := unisockets.Listen(serverSocket, 5); err != nil {
		return nil, err
	}

	return &TCPListener{
		fd:   serverSocket,
		addr: laddr,
	}, nil
}

type TCPListener struct {
	fd   int32
	addr net.Addr
}

func (t TCPListener) Close() error {
	return unisockets.Shutdown(t.fd, unisockets.SHUT_RDWR)
}

func (t TCPListener) Addr() net.Addr {
	return t.addr
}

func (l TCPListener) Accept() (net.Conn, error) {
	conn, err := l.AcceptTCP()

	return conn, err
}

func (l *TCPListener) AcceptTCP() (*TCPConn, error) {
	clientAddress := unisockets.SockaddrIn{}

	// Accept
	clientSocket, err := unisockets.Accept(l.fd, &clientAddress)
	if err != nil {
		return nil, err
	}

	return &TCPConn{
		fd: clientSocket,
		laddr: &TCPAddr{
			stringAddr: "",
			IP:         l.addr.(*TCPAddr).IP,
			Port:       l.addr.(*TCPAddr).Port,
			Zone:       "",
		},
		raddr: &TCPAddr{
			stringAddr: "",
			IP:         IP{byte(clientAddress.SinAddr.SAddr), byte(clientAddress.SinAddr.SAddr >> 8), byte(clientAddress.SinAddr.SAddr >> 16), byte(clientAddress.SinAddr.SAddr >> 24)},
			Port:       int(clientAddress.SinPort),
			Zone:       "",
		},
	}, nil
}

func Dial(network, address string) (net.Conn, error) {
	raddr, err := ResolveTCPAddr(network, address)
	if err != nil {
		return TCPConn{}, err
	}

	conn, err := DialTCP(network, nil, raddr) // TODO: Set laddr here
	if err != nil {
		return TCPConn{}, err
	}

	return *conn, err
}

func DialTCP(network string, laddr, raddr *TCPAddr) (*TCPConn, error) {
	// Create address
	serverAddress := unisockets.SockaddrIn{
		SinFamily: unisockets.PF_INET,
		SinPort:   unisockets.Htons(uint16(raddr.Port)),
		SinAddr: struct{ SAddr uint32 }{
			SAddr: binary.LittleEndian.Uint32(raddr.IP),
		},
	}

	// Create socket
	serverSocket, err := unisockets.Socket(unisockets.PF_INET, unisockets.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	// Connect
	if err := unisockets.Connect(serverSocket, &serverAddress); err != nil {
		return nil, err
	}

	return &TCPConn{
		fd:    serverSocket,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

type TCPConn struct {
	fd int32

	laddr net.Addr
	raddr net.Addr
}

func (c TCPConn) Read(b []byte) (int, error) {
	readMsg := make([]byte, len(b))

	n, err := unisockets.Recv(c.fd, &readMsg, uint32(len(b)), 0)
	if n == 0 {
		return int(n), errors.New("client disconnected")
	}

	copy(b, readMsg)

	return int(n), err
}

func (c TCPConn) Write(b []byte) (int, error) {
	n, err := unisockets.Send(c.fd, b, 0)
	if n == 0 {
		return int(n), errors.New("client disconnected")
	}

	return int(n), err
}

func (c TCPConn) Close() error {
	return unisockets.Shutdown(c.fd, unisockets.SHUT_RDWR)
}

func (c TCPConn) LocalAddr() net.Addr {
	return c.laddr
}

func (c TCPConn) RemoteAddr() net.Addr {
	return c.laddr
}

func (c TCPConn) SetDeadline(t time.Time) error {
	// TODO: Currently there is an infinite deadline

	return nil
}

func (c TCPConn) SetReadDeadline(t time.Time) error {
	// TODO: Currently there is an infinite deadline

	return nil
}

func (c TCPConn) SetWriteDeadline(t time.Time) error {
	// TODO: Currently there is an infinite deadline

	return nil
}
