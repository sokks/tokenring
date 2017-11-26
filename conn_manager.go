package tokenring

import (
	"time"
	"net"
)

type ConnManager struct {
	udpConn *net.UDPConn
	timeout time.Duration
	Fault   chan struct{}
	Token   chan TokenMessage
	kill    chan struct{}
}

func NewConnManager(port int, faultTimeout int) (*ConnManager) {
	bind := func(port) (*net.UDPConn) {
		laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			logger.Printf("ERROR: cannot resolve service address 127.0.0.1:%d\n", port)
		}
		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			logger.Println("ERROR: cannot bind service port", port)
		}
		return conn
	}

	m := &ConnManager{
		Fault: make(chan struct{}),
		Token: make(chan TokenMessage),
		kill:  make(chan struct{}),
	}

	m.udpConn = bind(port)
	m.timeout = time.Duration(faultTimeout * time.Millisecond)
	return m
} 

func (m *ConnManager) Start() {
	go m.work()
}

func (m *ConnManager) Stop() {
	unbind := func() {
		m.udpConn.Close()
	}
	m.kill <- struct{}{}
	time.Sleep(time.Duration(50 * time.Millisecond))
	unbind()
}

func (m *ConnManager) work() {
	buffer := make([]byte, 1024)
	for {
		select {
		case <-m.kill:
			return
		default:
			msg := TokenMessage{}
			// set timeout for token ring fault handling
			node.udpConn.SetReadDeadline(time.Now().Add(m.timeout))
			n, _, err := node.udpConn.ReadFromUDP(buffer)
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
					m.Token <- msg
				}
			}
		}
	}
}

func (m *ConnManager) Send(msg TokenMessage, port int) {
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		logger.Printf("ERROR: cannot resolve address 127.0.0.1:%d\n", port)
		return
	}
	buffer, _ := json.Marshal(msg)
	_, err = s.udpConn.WriteToUDP(buffer, raddr)
	if err != nil {
		_ = err
	}
} 