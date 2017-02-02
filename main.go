package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/peerdiscovery"
)

// Command line defaults
const (
	defaultJoinPort = ":11000"
	defaultRaftPort = ":12000"
)

// Command line parameters
var nick string
var simPort string
var raftPort string
var joinPort string

func main() {
	initLogger()
	// Parse arg flags
	flag.StringVar(&nick, "nick", "", "Nickname of this peer")
	flag.StringVar(&simPort, "sim", "", "Listening port of the simulator")
	flag.StringVar(&joinPort, "joinPort", defaultJoinPort, "Set Raft join port")
	flag.StringVar(&raftPort, "raftPort", defaultRaftPort, "Set Raft port")
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

	// Initialize peer discovery
	peers := make(map[string]peerdiscovery.Peer)
	// peerCh := make(chan string)
	go peerdiscovery.Start(33324, ownID.Nick, joinPort, raftPort, func(p peerdiscovery.Peer) {
		peers[p.Nick+"@"+p.IP] = p
	})
	time.Sleep(1 * time.Second)

	// Initialize driver
	cfg := driver.Config{
		SimMode: true,
		SimPort: "53566",
		Floors:  4,
		OnBtnPress: func(b driver.Btn) {
			driver.BtnLEDSet(b)
			time.Sleep(5 * time.Second)
			driver.BtnLEDClear(b)
		},
	}
	go driver.Init(cfg)

	// Initalize globalstate
	if len(peers) == 0 {
		go globalstate.Init("", joinPort, raftPort)
	} else {
		for _, anyPeer := range peers {
			go globalstate.Init(anyPeer.IP+":"+anyPeer.JoinPort, joinPort, raftPort)
			continue
		}
	}

	select {}

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
