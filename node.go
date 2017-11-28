package tokenring

import (
	"os"
	"log"
	"time"
	"strconv"
)

// Node contains internal info and works independently
type Node struct {
	ON          bool             // is switched on
	ringSize    int              // number of nodes in the net
	id          int              // node's id
	dataPort    int              // node's port number
	servicePort int              // node's maintainance port
	delay       time.Duration    // max marker hold time
	manager     *ConnManager     // node's token connection manager
	service     *NodeService     // node's maintainance connection receiver
	messages    []message        // messages waiting to be sent
	logger      *log.Logger      // node's logger with prefix
}

// TODO port number is not so good idea, because in process it is 
// counted as next (30000 + nextId): parameter next port would be better.

// NewNode conatructs new active node. Parameters are size of the ring, 
// node's id, token port, maintainance port, marker hold time.
func NewNode(isAM bool, size, nID, dPort, sPort, maxDelay int) (*Node) {
	return &Node{
		ON:          true,
		ringSize:    size,
		id:          nID,
		dataPort:    dPort,
		servicePort: sPort,
		delay:       time.Duration(maxDelay * int(time.Millisecond)),
		manager:     NewConnManager(dPort, maxDelay * (size + 2)),
		service:     NewNodeService(sPort),
		messages:    make([]message, 0, 10),
		logger:      log.New(os.Stdout, "[node " + strconv.Itoa(nID) + "] ", log.Ltime),
	}
}


func (node *Node) process(kill chan struct{}) {
	var (
		toDrop = 0
		wasTimeout = false
		toRecover = false
		failedNode = -1
		nextID = (node.id + 1) % node.ringSize
		toSend = make([]Token, 0, 10)
		//tt <-chan time.Time
	)
	processServiceMsg := func(msg ServiceMessage) {
		if msg.MsgType == "send" {
			newMsg := newMessage(msg.Dst, msg.Data)
			node.messages = append(node.messages, newMsg)
		} else if msg.MsgType == "terminate" {
			node.ON = false // TODO: process it when sending message; or AM should kill this message
			node.logger.Println("terminated")
		} else if msg.MsgType == "recover" {
			node.ON = true
			node.logger.Println("recovered")
			toRecover = true
		} else if msg.MsgType == "drop" {
			toDrop ++
		}
	}

	processToken := func(inMsg Token) (outMsg Token) {
		if inMsg.Free {
			if len(node.messages) > 0 {
				if node.messages[0].dst == failedNode {
					node.messages = append(node.messages[1:], node.messages[0])
					outMsg = NewEmptyToken(node.id)
				}
				outMsg = NewTokenWithMessage(node.id, node.messages[0])
			} else {
				outMsg = NewEmptyToken(node.id)
			}
		} else if inMsg.Fail {
			if failedID, _ := strconv.Atoi(inMsg.Data); failedID == nextID {
				failedNode = failedID
				nextID = (failedID + 1) % node.ringSize
				outMsg = NewEmptyToken(node.id)
			} else {
				failedNode = failedID
				outMsg = CopyToken(inMsg, node.id)
			}
		} else if inMsg.New {
			if newID, _ := strconv.Atoi(inMsg.Data); (newID > node.id) && (newID < nextID) {
				failedNode = -1
				nextID = newID
				outMsg = NewEmptyToken(node.id)
			} else {
				failedNode = -1
				outMsg = CopyToken(inMsg, node.id)
			}
		} else if inMsg.Dst == node.id {
			if inMsg.Ack {
				node.messages = node.messages[1:]
				outMsg = NewEmptyToken(node.id) // send empty to prevent monopolia
			} else {
				outMsg = NewAckToken(node.id, inMsg.Src)
			}
		} else {
			outMsg = CopyToken(inMsg, node.id)
		}
		return
	}
	
	node.manager.Start()
	defer node.logger.Println("manager terminated")
	defer time.Sleep(time.Duration((node.ringSize * 2) * int(node.delay)))
	defer node.manager.Stop()
	node.service.Start()
	defer node.logger.Println("service terminated")
	defer node.service.Stop()
	
	ticker := time.NewTicker(time.Duration(int(node.delay) * node.ringSize))
	defer ticker.Stop()

	for {
		select {
		case <-kill:
			node.logger.Printf("node termination")
			return
		case msg := <-node.service.C:
			node.logger.Println("received service message:", msg)
			processServiceMsg(msg)
		case msg := <-node.manager.Data:
			// how to process double receive??
			wasTimeout = false
			if !node.ON {
				continue
			} else if toDrop > 0 {
				toDrop--
				node.logger.Println("dropped token")
				continue
			} else {
				t := time.After(node.delay)
				newMsg := processToken(msg)
				node.logger.Printf("received %s, sending token to node %d\n", msg, nextID)
				toSend = append(toSend, newMsg)
				select {
				case <-t:
					node.manager.Send(newMsg, BasePort + nextID)
				}
			}
		case <-ticker.C: // only one token control
			if len(toSend) > 1 {
				toDrop ++
			}
			toSend = make([]Token, 0, 10)
		case <-node.manager.Fault:
			if !node.ON {
				continue
			} else if wasTimeout {
				// previous node terminated
				newMsg := NewInfoToken(node.id, (node.id + node.ringSize - 1) % node.ringSize, true)
				node.logger.Printf("fault detected, sending %s to node %d\n", newMsg, nextID)
				node.manager.Send(newMsg, BasePort + nextID)
			} else {
				// lost token
				wasTimeout = true
				node.logger.Printf("token timeout, sending new empty token to node %d\n", nextID)
				node.manager.Send(NewEmptyToken(node.id), BasePort + nextID)
			}
		}
		if toRecover {
			toRecover = false
			newMsg := NewInfoToken(node.id, node.id, false)
			node.logger.Printf("sending %s to node %d\n", newMsg, nextID)
			node.manager.Send(newMsg, BasePort + nextID)
		}
	}
}
