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
		size: n,
		kill: make(chan struct{}),
		delay: time.Duration(dl * time.Millisecond)
		nodes: make([]*Node, n)
	}
	for i := 0; i < n; i++ {
		nodes[i] = NewNode(n, i, BasePort + i, BaseServicePort + i)
	}
	return r
}

func (r *TokenRing) Start() {
	logger = initLogger()
	logger.Println(time.Now().String(), "Start")
	for i := 0; i < r.size; i++ {
		go r.nodes[i].process(r.kill, r.delay)
	}
	time.Sleep(time.Second)
}

func (r *TokenRing) Stop() {
	for i := 0; i <= GN.size; i++ {
		GN.kill <- struct{}{}
		time.Sleep(50 * time.Millisecond)
	}
	closeLogger()
}

func initLogger() (*log.Logger) {
	return log.New(os.Stdout, "Common:", log.Ltime)
}

func closeLogger() {
	// ??
}
	