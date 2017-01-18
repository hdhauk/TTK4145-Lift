package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/hw"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/network"
)

const (
	on  = true
	off = false
)

func main() {
	initLogger()
	// Handle application command-line flags
	var id string
	var simPort string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&simPort, "sim", "", "listening port of the simulator")
	flag.Parse()
	if simPort != "" {
		logger.Notice("Starting in simulator mode.")
	}
	id = peerName(id)

	// Setting up communication channels
	peerUpdateCh := make(chan network.PeerUpdate)
	peerTxEnable := make(chan bool)

	// Setting up running routines
	go network.HeartBeatBeacon(33324, id, peerTxEnable)
	go network.PeerMonitor(33324, peerUpdateCh)

	go hw.Init(simPort, logger)

	for {
		select {
		case p := <-peerUpdateCh:
			logPeerUpdate(p)

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
