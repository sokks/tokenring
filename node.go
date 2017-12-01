package tokenring

import (
	"os"
	"log"
	"time"
	"strconv"
)

type info struct {
	id    int
	sent  bool
	acked bool
}

// Node contains internal info and works independently
type Node struct {
	ON          bool             // is switched on
	ringSize    int              // number of nodes in the net
	id          int              // node's id
	delay       time.Duration    // max marker hold time
	manager     *ConnManager     // node's token connection manager
	service     *NodeService     // node's maintainance connection receiver
	messages    []Token          // messages waiting to be sent
	logger      *log.Logger      // node's logger with prefix
}

// NewNode conatructs new active node. Parameters are size of the ring, 
// node's id, token port, maintainance port, marker hold time.
func NewNode(size, nID, dPort, sPort, maxDelay int) (*Node) {
	return &Node{
		ON:          true,
		ringSize:    size,
		id:          nID,
		delay:       time.Duration(maxDelay * int(time.Millisecond)),
		manager:     NewConnManager(dPort, maxDelay * (size + 2)),
		service:     NewNodeService(sPort),
		messages:    make([]Token, 0, 10),
		logger:      log.New(os.Stdout, "[node " + strconv.Itoa(nID) + "] ", log.Ltime),
	}
}


func (node *Node) process(kill chan struct{}) {
	const (
		failedTimeouts = 3
		lossTimeouts   = 1
	)
	var (
		failedNode = -1    // currently terminated node's id (-1 for all alive)
		nextID     = (node.id + 1) % node.ringSize

		toDrop     = 0     // number of tokens to drop     
		toRecover  = false // flag of recovery (set with service message)
		
		timeouts   = 0     // number of timeouts occured in a row
		failure    *info   // notification info for other nodes
		recovery   *info   // ~ failure
	)

/****************************************************************************/	
/************************ service message processing  ***********************/	
/****************************************************************************/	

	processServiceMsg := func(msg ServiceMessage) {
		if msg.MsgType == "send" {
			newMsg := NewDataToken(node.id, msg.Dst, node.id, msg.Data)
			node.messages = append(node.messages, newMsg)
		} else if msg.MsgType == "terminate" {
			if node.ON {
				node.ON = false
				failedNode = node.id
				node.manager.Stop()
				node.logger.Println("terminated")
			}
		} else if msg.MsgType == "recover" {
			if !node.ON {
				node.ON = true
				node.manager.Start()
				node.logger.Println("recovered")
				toRecover = true
			}
		} else if msg.MsgType == "drop" {
			toDrop ++
		}
	}

/****************************************************************************/	
/********************** one token processing (complex) **********************/	
/****************************************************************************/	

	processToken := func(inMsg Token) (outMsg Token, notToSend bool) {
		if !inMsg.New && failure != nil && inMsg.Sndr == failure.id {
			node.logger.Println("failure canceled")
			if failure.acked {
				failure = nil
				recovery = &info{inMsg.Sndr, false, false}
			}
			failure = nil
			failedNode = -1
			notToSend = true
			return

		}
		if inMsg.Free {  // empty token
			if failure != nil && !failure.acked {
				outMsg = NewInfoToken(node.id, failure.id, true)
			} else if recovery != nil && !recovery.sent {
				outMsg = NewInfoToken(node.id, recovery.id, false)
				recovery.sent = true
			} else if recovery != nil && !recovery.acked {
				outMsg = NewInfoToken(node.id, recovery.id, false)
			} else if len(node.messages) > 0 {
				if (failedNode != -1) && node.messages[0].Dst == failedNode {
					newMessages := make([]Token, len(node.messages))
					for i := 1; i < len(node.messages); i++ {
						newMessages[i - 1] = node.messages[i]
					}
					newMessages[len(node.messages) - 1] = node.messages[0]
					node.messages = newMessages
					outMsg = NewEmptyToken(node.id)
				} else {
					outMsg = node.messages[0]
				}
			} else {
				outMsg = NewEmptyToken(node.id)
			}
		}  else if inMsg.Dst == node.id { // message to this node
			if inMsg.Ack {
				if len(node.messages) == 1 {
					node.messages = make([]Token, 0, 10)
				} else {
					node.messages = node.messages[1:]
				}
				outMsg = NewEmptyToken(node.id) // send empty to prevent monopolia
			} else {
				outMsg = NewAckToken(node.id, inMsg.Src)
			}
		} else if inMsg.Fail { // termination info
			failedID, _ := strconv.Atoi(inMsg.Data)
			if failure != nil && failedID == failure.id {
				failure.acked = true
				node.logger.Println("failure acked")
				outMsg = NewEmptyToken(node.id)
				return
			}
			if failure == nil && recovery != nil && failedID == recovery.id {
				node.logger.Println("failure acked")
				outMsg = NewEmptyToken(node.id)
				return
			}
			if failure != nil && failedID != failure.id {
				failure = nil
			}
			failedNode = failedID
			if failedID == nextID {
				nextID = (failedID + 1) % node.ringSize
			}
			outMsg = CopyToken(inMsg, node.id)
		} else if inMsg.New {  // recovery info
			newID, _ := strconv.Atoi(inMsg.Data)
			if failure != nil && inMsg.Sndr == failure.id && newID == failure.id {
				node.logger.Println("recovery detected")
				failure = nil
				recovery = &info{inMsg.Sndr, false, false}
				failedNode = -1
				notToSend = true
				return
			}
			
			if recovery != nil && recovery.id == newID { // recovery acked
				recovery = nil
				outMsg = NewEmptyToken(node.id)
				return
			}
			failedNode = -1
			if (node.id + 1) % node.ringSize == newID {
				nextID = newID
			}
			outMsg = CopyToken(inMsg, node.id)
		} else {
			outMsg = CopyToken(inMsg, node.id)
		}
		return
	}

/****************************************************************************/	
/******************************* MAIN PROCESS *******************************/	
/****************************************************************************/	

	node.manager.Start()
	defer node.logger.Println("manager terminated")
	defer time.Sleep(time.Duration((node.ringSize * 2) * int(node.delay)))
	defer node.manager.Stop()
	node.service.Start()
	defer node.logger.Println("service terminated")
	defer node.service.Stop()

	for {
		select {
		case <-kill:
			node.logger.Printf("node termination")
			return
		case msg := <-node.service.C:
			node.logger.Println("received service message:", msg)
			processServiceMsg(msg)
		case msg := <-node.manager.Data:
			timeouts = 0
			if !node.ON {
				continue
			} else if toDrop > 0 {
				toDrop--
				node.logger.Println("dropped token")
				continue
			} else {
				t := time.After(node.delay) // one node delay control
				newMsg, notToSend := processToken(msg)
				if notToSend {
					node.logger.Printf("received %s\n", msg)
				} else {
					node.logger.Printf("received %s, sending token to node %d\n", msg, nextID)
					select {
					case <-t:
						node.manager.Send(newMsg, BasePort + nextID)
					}
				}
			}
		case <-node.manager.Fault: // timeout
			if !node.ON {
				continue
			} else {
				timeouts++
				node.logger.Println("timeouts =", timeouts)
				if (timeouts > failedTimeouts) && failedNode == -1 {
					// previous node terminated
					failedNode = (node.id + node.ringSize - 1) % node.ringSize
					failure = &info{failedNode, true, false}
					newMsg := NewInfoToken(node.id, failedNode, true)
					node.logger.Printf("termination detected, sending %s to node %d", newMsg, nextID)
					node.manager.Send(newMsg, BasePort + nextID)
				} else if timeouts > failedTimeouts && failure != nil {
					// previous node terminated, it is already detected but not corrected
					newMsg := NewInfoToken(node.id, failure.id, true)
					node.logger.Printf("termination detected again, sending %s to node %d", newMsg, nextID)
					node.manager.Send(newMsg, BasePort + nextID)
				} else if timeouts > lossTimeouts {
					// lost token
					node.logger.Printf("token timeout, sending new empty token to node %d\n", nextID)
					node.manager.Send(NewEmptyToken(node.id), BasePort + nextID)
				}
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
