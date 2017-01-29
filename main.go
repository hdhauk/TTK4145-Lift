package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/peerdiscovery"
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
	flag.StringVar(&nick, "nick", "", "nick name of this peer")
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
	// peerUpdateCh := make(chan peerdiscovery.PeerUpdate)
	//peerTxEnable := make(chan bool)

	// Setting up running routines
	//go peerdiscovery.HeartBeatBeacon(33324, ownID.Nick, peerTxEnable)
	go peerdiscovery.Start(33324, ownID.Nick, func(id, IP string) {
		fmt.Printf("ID = %v\n", id)
		fmt.Printf("IP = %v\n", IP)
	})

	//go driver.Init(true, simPort)

	for {
		select {
		case <-time.After(1 * time.Second):
			//logPeerUpdate(p)

		}
	}
}

func peerName(id string) string {
	if id == "" {
		id = strconv.Itoa(os.Getpid())
	}
	localIP, err := peerdiscovery.GetLocalIP()
	if err != nil {
		logger.Warning("Not connected to the internet.")
		localIP = "DISCONNECTED"
	}
	return fmt.Sprintf("%s@%s", id, localIP)
}
