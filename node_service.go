package tokenring

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"
)

type NodeService struct {
	C       chan ServiceMessage
	udpConn *net.UDPConn
	kill    chan struct{}
	buffer  []byte
}

func NewNodeService(port int) (*NodeService) {
	bind := func(port int) (*net.UDPConn) {
		laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			fmt.Printf("ERROR: cannot resolve service address 127.0.0.1:%d", port)
		}
		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
		}
		return conn
	}
	
	s := &NodeService{
		C:      make(chan ServiceMessage, 5),
		kill:   make(chan struct{}),
		buffer: make([]byte, 1024),
	}
	s.udpConn = bind(port)
	return s
}

func (s *NodeService) Start() {
	go s.listen()
}

func (s *NodeService) Stop() {
	unbind := func() {
		s.udpConn.Close()
	}
	s.kill <- struct{}{}
	time.Sleep(time.Duration(50 * time.Millisecond))
	unbind()
}

func (s *NodeService) listen() {
	timeout := time.Duration(50 * time.Millisecond)
	Loop:
	for {
		select {
		case <-s.kill:
			return
		default:
			msg := ServiceMessage{}
			// set timeout for non-blocking in case of no messages
			s.udpConn.SetReadDeadline(time.Now().Add(timeout))
			n, _, err := s.udpConn.ReadFromUDP(s.buffer)
			if err != nil {
				if e, ok := err.(net.Error); !ok || !e.Timeout() {
					panic(err)
				} else {
					continue Loop
				}
			} else {
				if n != 0 {
					json.Unmarshal(s.buffer[:n], &msg)
					s.C <- msg
				}
			}
		}
	}
}