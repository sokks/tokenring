package tokenring

import (
	"os"
	"log"
	"time"
	"strconv"
)

type Node struct {
	ringSize    int
	id          int
	dataPort    int
	servicePort int
	delay       time.Duration
	manager     *ConnManager
	service     *NodeService
	messages    []TokenMessage
	logger      *log.Logger
}

func NewNode(size, nId, dPort, sPort, maxDelay int) (*Node) {
	return &Node{
		ringSize:    size,
		id:          nId,
		dataPort:    dPort,
		servicePort: sPort,
		delay:       time.Duration(maxDelay * time.Millisecond),
		manager:     NewConnManager(dPort, maxDelay * (size + 5)),
		service:     NewNodeService(sPort),
		messages:    make([]TokenMessage, 0, 10),
		logger:      log.New(os.Stdout, "Node " + strconv.Itoa(nId) + ":"),
	}
}


func (node *Node) process(kill chan struct{}, delay time.Duration) {
	node.bind()
	defer node.unbind()
	node.service.Start()
	defer node.service.Stop()
	node.manager.Start()
	defer node.manager.Stop()
	timeout := time.Duration(int(delay) * (node.ringSize + 5))
	buffer := make([]byte, 1024)
	for {
		select {
		case <-kill:
			return
		case msg := <-node.service.C:
			node.logger.Println("received service message:", msg)
			// do what it requests
		case <- node.manager.Fault:
			// process fault
		case msg := <- node.manager.Token:
			// get type (empty/with data/with ack)
			node.logger.Printf("received token", msg)
			// set delay timer TODO more elegant
			t := time.After(delay)

			// process message

			for {
				select {
				case <-t:
					break
				}
			}
			// send further
		}
	}
}
