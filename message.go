package tokenring

import (
	"fmt"
	"strconv"
)

// ServiceMessage represents a message that is send to a service 
// port of node to control its behaviour. Possible types are 
// "send", "terminate", "recover", "drop".
type ServiceMessage struct {
	MsgType string `json:"type"`
	Dst     int    `json:"dst"`
	Data    string `json:"data"`
}

// NewServiceMessage creates new message from input parameters. 
// Possible types are "send", "terminate", "recover", "drop".
func NewServiceMessage(msgType string, dst int, data string) ServiceMessage {
	return ServiceMessage{ msgType, dst, data }
}

func (m ServiceMessage) String() string {
	return fmt.Sprintf("{type: %s, dst: %d, data: %s}", m.MsgType, m.Dst, m.Data)
}

type message struct {
	dst  int    // where to send
	data string // what to send
}

func newMessage(dst int, data string) message {
	return message{dst, data}
}


// TODO: maybe better unexported?

// Token represents an internal structure used by nodes
// to передача/передавать token from one to another.
type Token struct {
	// flags
	Free bool   `json:"free"`
	Ack  bool   `json:"ack"`
	Fail bool   `json:"fail"`
	New  bool   `json:"new"`
	// info
	Src  int    `json:"src"`
	Dst  int    `json:"dst"`
	Sndr int    `json:"sender"`
	// data
	Data string `json:"data"`
}
// Can omit sender if process it externally using ReceiveFromUDP. 
// It's here just to simplify processing.


// NewEmptyToken creates new message from input parameters.
func NewEmptyToken(sender int) Token {
	return Token{ true, false, false, false, -1, -1, sender, "" }
}

// NewDataToken create new message with data.
func NewDataToken(src, dst, sndr int, data string) Token {
	return Token{ false, false, false, false, src, dst, sndr, data }
}

// NewTokenWithMessage create new message with message.
func NewTokenWithMessage(src int, msg message) Token {
	return Token{ false, false, false, false, src, msg.dst, src, msg.data }
}

// NewAckToken creates new message with delivery confirmation.
func NewAckToken(src, dst int) Token {
	return Token{ false, true, false, false, src, dst, src, "" }
}

// NewInfoToken creates new message with delivery confirmation.
// isFail = true if it's info about node fail
// isFail = false if it's info about recovery
func NewInfoToken(src, infoID int, isFail bool) Token {
	return Token{ false, false, isFail, !isFail, src, -1, src, strconv.Itoa(infoID) }
}

// CopyToken creates a copy of token to propagate by sndr.
func CopyToken(t Token, sndr int) Token {
	return Token{ t.Free, t.Ack, t.Fail, t.New, t.Src, t.Dst, sndr, t.Data }
}

func (m Token) String() string {
	if m.Free {
		return fmt.Sprintf("empty token from node %d", m.Sndr)
	} else if m.Ack {
		return fmt.Sprintf("token from node %d with delivery confirmation from node %d to node %d", 
				m.Sndr, m.Src, m.Dst)
	} else if m.Fail {
		return fmt.Sprintf("token from node %d with fault info on node %s", m.Sndr, m.Data)
	} else if m.New {
		return fmt.Sprintf("token from node %d with recovery info on node %s", m.Sndr, m.Data)
	}
	return fmt.Sprintf("token from node %d with data from %d(data = \"%s\") to %d", 
			m.Sndr, m.Src, m.Data, m.Dst)
}