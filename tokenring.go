package tokenring

import (
	"time"
	"log"
	"os"
)

const (
	BasePort = 30000
	BaseServicePort = 40000
)

var (
	logger *log.Logger
)

type TokenRing struct {
	size  int
	kill  chan struct{}
	delay time.Duration
	nodes []*Node
}

func NewTokenRing(n, dl int) (*TokenRing) {
	r := &TokenRing{
		size:  n,
		kill:  make(chan struct{}),
		delay: time.Duration(dl * int(time.Millisecond)),
		nodes: make([]*Node, n),
	}
	for i := 0; i < n; i++ {
		r.nodes[i] = NewNode(n, i, BasePort + i, BaseServicePort + i, dl)
	}
	return r
}

func (r *TokenRing) Start() {
	logger = initLogger()
	logger.Println(time.Now().String(), "Start")
	for i := 0; i < r.size; i++ {
		go r.nodes[i].process(r.kill)
	}
	time.Sleep(time.Second)
	r.nodes[0].manager.Send(NewEmptyTokenMessage(0), BasePort + 1)
}

func (r *TokenRing) Stop() {
	for i := 0; i < r.size; i++ {
		r.kill <- struct{}{}
		time.Sleep(50 * time.Millisecond)
	}
	logger.Println(time.Now().String(), "Stop")
}

func initLogger() (*log.Logger) {
	return log.New(os.Stdout, "Common:", log.Ltime)
}
