package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

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

// Pick ports randomly
var r = 1024 + rand.Intn(64510)
var raftPort = strconv.Itoa(r)
var commPort = strconv.Itoa(r + 1)

func main() {
	fmt.Printf("raftPort: %s, joinPort: %s\n", raftPort, commPort)
	initLogger()
	// Parse arg flags
	flag.StringVar(&nick, "nick", "", "Nickname of this peer")
	flag.StringVar(&simPort, "sim", "", "Listening port of the simulator")

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
	go peerdiscovery.Start(33324, ownID.Nick, commPort, raftPort, func(p peerdiscovery.Peer) {
		peers[p.Nick+"@"+p.IP] = p
	})
	time.Sleep(1 * time.Second)

	// Initialize driver
	// cfg := driver.Config{
	// 	SimMode: true,
	// 	SimPort: "53566",
	// 	Floors:  4,
	// 	OnBtnPress: func(b driver.Btn) {
	// 		driver.BtnLEDSet(b)
	// 		time.Sleep(5 * time.Second)
	// 		driver.BtnLEDClear(b)
	// 	},
	// }
	// go driver.Init(cfg)

	// Initalize globalstate
	ip, _ := peerdiscovery.GetLocalIP()
	gsCfg := globalstate.Config{
		RaftPort: r,
		OwnIP:    ip,
	}

	if len(peers) == 0 {
		go globalstate.Init(gsCfg)
	} else {
		for _, anyPeer := range peers {
			logger.Noticef("Identified raft. Starting global state with connection to: %s\n", anyPeer.IP+":"+anyPeer.JoinPort)
			gsCfg.InitalPeer = anyPeer.IP + ":" + anyPeer.JoinPort
			go globalstate.Init(gsCfg)
			break
		}
	}

	time.Sleep(5 * time.Second)
	status := globalstate.LiftStatusUpdate{
		Floor: 1,
		Dst:   2,
		Dir:   "DOWN",
	}
	globalstate.UpdateLiftStatus(status)

	bsu := globalstate.ButtonStatusUpdate{
		Floor:  2,
		Dir:    "down",
		Status: globalstate.BtnStateDone,
	}
	globalstate.UpdateButtonStatus(bsu)

	time.Sleep(10 * time.Second)
	temp := globalstate.GetState()
	fmt.Printf("%+v", temp)

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
