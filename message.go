package tokenring

import "fmt"

// ServiceMessage represents a message that is send to a service 
// port of node to control its behaviour. Possible types are 
// "send", "terminate", "recover", "drop".
type ServiceMessage struct {
	MsgType string `json:"type"`
	Dst     int    `json:"dst"`
	Data    string `json:"data"`
}

// NewServiceMessage creates new message from input parameters.
func NewServiceMessage(msgType string, dst int, data string) ServiceMessage {
	return ServiceMessage{ msgType, dst, data }
}

func (m ServiceMessage) String() string {
	return fmt.Sprintf("{type: %s, dst: %d, data: %s}", m.MsgType, m.Dst, m.Data)
}

// TODO: maybe better unexported?

// TokenMessage represents an internal structure used by nodes
// to передача/передавать token from one to another.
type TokenMessage struct {
	Free bool   `json:"free"`
	Src  int    `json:"src"`
	Dst  int    `json:"dst"`
	Sndr int    `json:"sender"`
	Data string `json:"data"`
	Ack  bool   `json:"ack"`
}
// Can omit sender if process it externally using ReceiveFromUDP. 
// It's here just to simplify processing.


// NewEmptyTokenMessage creates new message from input parameters.
func NewEmptyTokenMessage(sender int) TokenMessage {
	return TokenMessage{ true, -1, -1, sender, "", false }
}

// NewTokenMessage creates new message from input parameters.
func NewTokenMessage(src int, dst int, sndr int, data string, isAck bool) TokenMessage {
	return TokenMessage{ false, src, dst, sndr, data, isAck }
}

func (m TokenMessage) String() string {
	if m.Free {
		return fmt.Sprintf("empty token from node %d", m.Sndr)
	} else if m.Ack {
		return fmt.Sprintf("token from node %d with delivery confirmation from node %d to node %d", 
				m.Sndr, m.Src, m.Dst)
	} 
	return fmt.Sprintf("token from node %d with data from %d(data = \"%s\") to %d", 
			m.Sndr, m.Src, m.Data, m.Dst)
}