package tokenring

import (
	"fmt"
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
		manager:     NewConnManager(dPort, maxDelay * (size + 2)),
		service:     NewNodeService(sPort),
		messages:    make([]TokenMessage, 0, 10),
		logger:      log.New(os.Stdout, "[node " + strconv.Itoa(nId) + "] ", log.Ltime),
	}
}


func (node *Node) process(kill chan struct{}) {
	var (
		toDrop = false
	)
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
			toDrop = true
			fmt.Println(toDrop)
		}
	}

	processToken := func(inMsg TokenMessage) (outMsg TokenMessage) {
		if inMsg.Free {
			if len(node.messages) > 0 {
				outMsg = node.messages[0]
			} else {
				outMsg = NewEmptyTokenMessage(node.id)
			}
		} else if inMsg.Dst == node.id {
			if inMsg.Ack {
				node.messages = node.messages[1:]
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
	
	node.manager.Start()
	defer node.logger.Println("manager terminated")
	defer time.Sleep(time.Duration((node.ringSize * 2) * int(node.delay)))
	defer node.manager.Stop()
	node.service.Start()
	defer node.logger.Println("service terminated")
	defer node.service.Stop()
	//var lastSent TokenMessage
	for {
		select {
		case <-kill:
			node.logger.Printf("node termination")
			return
		case msg := <-node.service.C:
			node.logger.Println("received service message:", msg)
			processServiceMsg(msg)
		case msg := <-node.manager.Token:
			if toDrop {
				toDrop = false
				node.logger.Println(toDrop)
				continue
			} else {
				t := time.After(node.delay)
				newMsg := processToken(msg)
				node.logger.Printf("received %s, sending token to node %d\n", msg, getNextId())
				select {
				case <-t:
					node.manager.Send(newMsg, BasePort + getNextId())
					//lastSent = newMsg
				}
			}
		case <-node.manager.Fault:
			// process fault TODO
			node.logger.Printf("token timeout, sending empty token to node %d\n", getNextId())
			node.manager.Send(NewEmptyTokenMessage(node.id), BasePort + getNextId())
		}
	}
}
