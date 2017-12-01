package tokenring

import (
	"encoding/json"
	"net"
	"strconv"
	"time"
)

// ConnManager handles socket operations. It sends all received data 
// to Data channel and all timeouts information to Fault channel.
type ConnManager struct {
	addr    *net.UDPAddr
	udpConn *net.UDPConn
	timeout time.Duration
	Fault   chan struct{}
	Data    chan Token
	kill    chan struct{}
}

// NewConnManager creates UDP connection manager for port with faultTimeout.
// Constructor does not bind port. Use Start() to start listening.
func NewConnManager(port int, faultTimeout int) (*ConnManager) {
	laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		logger.Printf("ERROR: cannot resolve service address 127.0.0.1:%d\n", port)
	}

	m := &ConnManager{
		addr:  laddr,
		Fault: make(chan struct{}, 2),
		Data:  make(chan Token, 2),
		kill:  make(chan struct{}),
	}

	m.timeout = time.Duration(faultTimeout * int(time.Millisecond))
	return m
} 

// Start opens UDP connection and starts listening in seporate goroutine.
func (m *ConnManager) Start() {
	conn, err := net.ListenUDP("udp", m.addr)
	if err != nil {
		logger.Println("ERROR: cannot bind address", m.addr)
	}
	m.udpConn = conn
	time.Sleep(10 * time.Millisecond)
	go m.work()
}

// Stop closes connection and stops listening goroutine.
func (m *ConnManager) Stop() {
	m.kill <- struct{}{}
	time.Sleep(time.Duration(10 * time.Millisecond))
	m.udpConn.Close()
}

func (m *ConnManager) work() {
	buffer := make([]byte, 1024)
	for {
		select {
		case <-m.kill:
			return
		default:
			msg := Token{}
			// set timeout for token ring fault handling
			m.udpConn.SetReadDeadline(time.Now().Add(m.timeout))
			n, _, err := m.udpConn.ReadFromUDP(buffer)
			if err != nil {
				if e, ok := err.(net.Error); !ok || !e.Timeout() {
					panic(err)
				} else {
					// process fault (no token)
					m.Fault <- struct{}{}
				}
			} else {
				if n != 0 {
					json.Unmarshal(buffer[:n], &msg)
					// process token
					m.Data <- msg
				}
			}
		}
	}
}

// Send engages connection manager to send message to port of lo.
func (m *ConnManager) Send(msg Token, port int) {
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		logger.Printf("ERROR: cannot resolve address 127.0.0.1:%d\n", port)
		return
	}
	buffer, _ := json.Marshal(msg)
	_, err = m.udpConn.WriteToUDP(buffer, raddr)
	if err != nil {
		logger.Println(err)
	}
} 