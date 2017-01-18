package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/hw"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/network"
)

func peerName(id string) string {
	if id == "" {
		id = strconv.Itoa(os.Getpid())
	}
	localIP, err := network.LocalIP()
	if err != nil {
		log.Warning("Not connected to the internet.")
		localIP = "DISCONNECTED"
	}
	return fmt.Sprintf("%s@%s", id, localIP)
}

const (
	on  = true
	off = false
)

func main() {
	initLogger()
	// Handle application flags
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	isSim := flag.Bool("sim", false, "simulator mode")
	flag.Parse()
	if *isSim {
		log.Notice("Starting in simulator mode.")
	}
	id = peerName(id)

	peerUpdateCh := make(chan network.PeerUpdate)
	// We can disable/enable the 	hw "bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	// transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go network.Transmitter(53566, id, peerTxEnable)
	go network.Receiver(53566, peerUpdateCh)

	go hw.Init(true)

	time.Sleep(1 * time.Second)
	hw.SetBtnLED(hw.Btn{Floor: 0, Type: hw.HallUp}, on)
	hw.SetDoorLED(off)
	time.Sleep(2 * time.Second)
	hw.SetDoorLED(on)

	for {
		select {
		case p := <-peerUpdateCh:
			logPeerUpdate(p)

		}
	}
}
