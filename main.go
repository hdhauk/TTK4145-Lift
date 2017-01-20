package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/hw"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/msg"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/network"
)

const (
	on  = true
	off = false
)

func main() {
	initLogger()
	// Handle application command-line flags
	var nick string
	var simPort string
	flag.StringVar(&nick, "nick", "", "nick name name of this peer")
	flag.StringVar(&simPort, "sim", "", "listening port of the simulator")
	flag.Parse()
	if simPort != "" {
		logger.Notice("Starting in simulator mode.")
	}
	var ownID = struct {
		ID   string
		Nick string
	}{
		ID: makeUUID(), Nick: peerName(nick),
	}

	// Setting up communication channels
	peerUpdateCh := make(chan network.PeerUpdate)
	peerTxEnable := make(chan bool)
	rxOrderCh := make(chan msg.Order)
	txOrderCh := make(chan msg.Order)

	// Setting up running routines
	go network.HeartBeatBeacon(33324, ownID.Nick, peerTxEnable)
	go network.PeerMonitor(33324, peerUpdateCh)
	go network.BcastReceiver(36969, rxOrderCh)
	go network.BcastTransmitter(36969, txOrderCh)
	go func(id string) {
		time.Sleep(3 * time.Second)
		fmt.Println("sending order")
		txOrderCh <- msg.Order{OrderID: makeUUID(), SrcID: id, Dir: "UP", Floor: 3}
	}(ownID.ID)

	go hw.Init(simPort, logger)

	for {
		select {
		case p := <-peerUpdateCh:
			logPeerUpdate(p)
		case o := <-rxOrderCh:
			fmt.Println(o)

		}
	}
}

func peerName(id string) string {
	if id == "" {
		id = strconv.Itoa(os.Getpid())
	}
	localIP, err := network.LocalIP()
	if err != nil {
		logger.Warning("Not connected to the internet.")
		localIP = "DISCONNECTED"
	}
	return fmt.Sprintf("%s@%s", id, localIP)
}
