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
		delay:       time.Duration(maxDelay * int(time.Millisecond)),
		manager:     NewConnManager(dPort, maxDelay * (size + 5)),
		service:     NewNodeService(sPort),
		messages:    make([]TokenMessage, 0, 10),
		logger:      log.New(os.Stdout, "Node " + strconv.Itoa(nId) + ":", log.Ltime),
	}
}


func (node *Node) process(kill chan struct{}) {
	processServiceMsg := func(msg ServiceMessage) {
		if msg.MsgType == "send" {
			newMsg := NewTokenMessage(node.id, msg.Dst, node.id, msg.Data, false)
			node.messages = append(node.messages, newMsg)
		} else if msg.MsgType == "terminate" {
			// switch off
		} else if msg.MsgType == "recover" {
			// switch on
		} else if msg.MsgType == "drop" {
			// drop one token
		}
	}

	processToken := func(inMsg TokenMessage) (outMsg TokenMessage) {
		if inMsg.Free {
			if len(node.messages) > 0 {
				outMsg = node.messages[0]
				node.messages = node.messages[1:] // check!
			} else {
				outMsg = NewEmptyTokenMessage(node.id)
			}
		} else if inMsg.Dst == node.id {
			if inMsg.Ack {
				outMsg = NewEmptyTokenMessage(node.id)
			} else {
				outMsg = NewTokenMessage(node.id, inMsg.Src, node.id, "", true)
			}
		} else {
			outMsg = NewTokenMessage(inMsg.Src, inMsg.Dst, node.id, inMsg.Data, inMsg.Ack)
		}
		return
	}

	getNextId := func() int { // TODO modify to get actual next (if fault)
		return (node.id + 1) % node.ringSize
	}
	
	node.service.Start()
	defer node.service.Stop()
	node.manager.Start()
	defer node.manager.Stop()
	var lastSent TokenMessage
	for {
		select {
		case <-kill:
			return
		case msg := <-node.service.C:
			node.logger.Println("received service message:", msg)
			processServiceMsg(msg)
		case <- node.manager.Fault:
			// process fault TODO
			t := time.After(node.delay)
			node.logger.Printf("token timeout, sending copy to node %d\n", getNextId())
			select {
			case <-t:
				node.manager.Send(lastSent, getNextId())
			}
		case msg := <-node.manager.Token:
			t := time.After(node.delay)
			newMsg := processToken(msg)
			node.logger.Printf("received %s, sending token to node %d\n", msg, getNextId())
			select {
			case <-t:
				node.manager.Send(newMsg, BasePort + getNextId())
				lastSent = newMsg
			}
		}
	}
}
