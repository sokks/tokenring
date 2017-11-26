package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/sokks/tokenring"
)

const (
	MainPort = 39999
)

func main() {
	fmt.Println(os.Args)
	n, _ := strconv.Atoi(os.Args[1])
	dl, _ := strconv.Atoi(os.Args[2])
	tr := tokenring.NewTokenRing(n, dl)
	tr.Start()
	fmt.Scanln()
	sendServiceMsg(tokenring.NewServiceMessage("send", 3, "lalala"), 40000)
	fmt.Scanln()
	sendServiceMsg(tokenring.NewServiceMessage("terminate", 1, ""), 40001)
	fmt.Scanln()
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