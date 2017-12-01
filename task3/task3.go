package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/sokks/tokenring"
)

const (
	MainPort = 39999
)

func main() {
	sizePtr := flag.Int("n", 5, "number of nodes in the token ring")
	delayPtr := flag.Int("t", 1000, "one node token delay (milliseconds)")
	flag.Parse()
	tr := tokenring.NewTokenRing(*sizePtr, *delayPtr)

	tr.Start()
	time.Sleep(2 * time.Second)

	sendServiceMsg(tokenring.NewServiceMessage("send", 3, "lalala1"), tokenring.BaseServicePort + 0)
//	sendServiceMsg(tokenring.NewServiceMessage("drop", 0, ""), tokenring.BaseServicePort + 2)
	sendServiceMsg(tokenring.NewServiceMessage("send", 3, "lalala2"), tokenring.BaseServicePort + 0)
	sendServiceMsg(tokenring.NewServiceMessage("send", 3, "lalala3"), tokenring.BaseServicePort + 1)
	sendServiceMsg(tokenring.NewServiceMessage("send", 3, "lalala4"), tokenring.BaseServicePort + 2)
	
	sendServiceMsg(tokenring.NewServiceMessage("terminate", 2, ""), tokenring.BaseServicePort + 2)

	time.Sleep(30 * time.Second)
	sendServiceMsg(tokenring.NewServiceMessage("drop", 0, ""), tokenring.BaseServicePort+ 0)
// 	sendServiceMsg(tokenring.NewServiceMessage("drop", 0, ""), tokenring.BaseServicePort+ 0)
// 	sendServiceMsg(tokenring.NewServiceMessage("drop", 0, ""), tokenring.BaseServicePort+ 0)
// 	sendServiceMsg(tokenring.NewServiceMessage("drop", 0, ""), tokenring.BaseServicePort+ 0)

//	fmt.Scanln()
	time.Sleep(1 * time.Minute)
	sendServiceMsg(tokenring.NewServiceMessage("recover", 2, ""), tokenring.BaseServicePort + 2)

	time.Sleep(30 * time.Second)
	sendServiceMsg(tokenring.NewServiceMessage("terminate", 4, ""), tokenring.BaseServicePort + 4)
	time.Sleep(30 * time.Second)
	sendServiceMsg(tokenring.NewServiceMessage("recover", 4, ""), tokenring.BaseServicePort + 4)

	time.Sleep(10 * time.Second)
//	fmt.Scanln()
	tr.Stop()
}

func sendServiceMsg(msg tokenring.ServiceMessage, port int) {
	laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(MainPort)))
	if err != nil {
		fmt.Printf("ERROR: cannot resolve address 127.0.0.1:%d\n", MainPort)
	}
	udpConn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		fmt.Println("ERROR: cannot bind port", MainPort)
	}
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		fmt.Printf("ERROR: cannot resolve address 127.0.0.1:%d\n", port)
		return
	}
	buffer, _ := json.Marshal(msg)
	_, err = udpConn.WriteToUDP(buffer, raddr)
	udpConn.Close()
}