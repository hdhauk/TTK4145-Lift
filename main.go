package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
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

// Pick ports randomly
var r = 1024 + rand.Intn(64510)
var raftPort = r

func main() {
	mainlogger := log.New(os.Stderr, "[main] ", log.Ltime|log.Lshortfile)
	// Parse arg flags
	flag.StringVar(&nick, "nick", strconv.Itoa(os.Getpid()), "Nickname of this peer. Default is the process id (PID)")
	flag.StringVar(&simPort, "sim", "", "Listening port of the simulator")
	flag.IntVar(&raftPort, "raft", raftPort, "Communication port for raft")
	flag.Parse()
	fmt.Printf("raftPort: %d, nick=%s, simulator=%s\n", raftPort, nick, simPort)

	// Initialize peer discovery
	peers := make(map[string]peerdiscovery.Peer)
	discoveryConfig := peerdiscovery.Config{
		Nick:              nick,
		RaftPort:          raftPort,
		BroadcastPort:     33324,
		OnNewPeer:         func(p peerdiscovery.Peer) { peers[p.Nick+":"+p.IP] = p },
		OnLostPeer:        onLostPeer,
		BroadcastInterval: 15 * time.Millisecond,
		Timeout:           50 * time.Millisecond,
		Logger:            log.New(os.Stderr, "[peerdiscovery] ", log.Ltime|log.Lshortfile),
	}
	go peerdiscovery.Start(discoveryConfig)
	time.Sleep(2 * discoveryConfig.BroadcastInterval)

	// Initialize driver
	cfg := driver.Config{
		SimMode:    true,
		SimPort:    simPort,
		Floors:     4,
		OnBtnPress: onBtnPress,
		Logger:     log.New(os.Stderr, "[driver] ", log.Ltime|log.Lshortfile),
	}
	driverInitDone := make(chan struct{})
	go driver.Init(cfg, driverInitDone)
	mainlogger.Println("[INFO] Waiting for driver to initialize")
	<-driverInitDone
	mainlogger.Println("[INFO] Driver ready")

	// Initalize globalstate
	ip, _ := peerdiscovery.GetLocalIP()
	globalstateConfig := globalstate.Config{
		RaftPort:          raftPort,
		OwnIP:             ip,
		OnPromotion:       func() {},
		OnDemotion:        func() { fmt.Println("DEMOTED, :(") },
		OnIncomingCommand: onIncommingCommand,
		CostFunction:      func(s globalstate.State, f int, d string) string { return "localhost:8003" },
		Logger:            log.New(os.Stderr, "[globalstate] ", log.Ltime|log.Lshortfile),
	}

	// Pass any known peers
	if len(peers) == 0 {
		go globalstate.Init(globalstateConfig)
	} else {
		for _, anyPeer := range peers {
			globalstateConfig.InitalPeer = anyPeer.IP + ":" + anyPeer.RaftPort
			mainlogger.Printf("Identified raft. Starting global state with connection to: %s\n", globalstateConfig.InitalPeer)
			go globalstate.Init(globalstateConfig)
			break
		}
	}

	time.Sleep(5 * time.Second)
	status := globalstate.LiftStatusUpdate{
		Floor: 1,
		Dst:   2,
		Dir:   "down",
	}
	globalstate.UpdateLiftStatus(status)

	bsu := globalstate.ButtonStatusUpdate{
		Floor:  3,
		Dir:    "down",
		Status: globalstate.BtnStateUnassigned,
	}
	globalstate.UpdateButtonStatus(bsu)

	time.Sleep(10 * time.Second)
	temp, _ := globalstate.GetState()
	fmt.Printf("%+v", temp)

	select {}

}

func peerName(id string) string {
	if id == "" {
		id = strconv.Itoa(os.Getpid())
	}
	localIP, err := peerdiscovery.GetLocalIP()
	if err != nil {
		localIP = "DISCONNECTED"
	}
	return fmt.Sprintf("%s:%s", id, localIP)
}
