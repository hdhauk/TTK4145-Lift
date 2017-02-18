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
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/statetools"
)

// Command line parameters
var nick string
var simPort string

// Pick ports randomly
var r = 1024 + rand.Intn(64510)
var raftPort = r

var gs globalstate.FSM

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
		SimMode: false,
		//SimPort: simPort,
		Floors:       4,
		OnBtnPress:   onBtnPress,
		OnNewStatus:  onNewStatus,
		OnDstReached: onDstReached,
		Logger:       log.New(os.Stderr, "[driver] ", log.Ltime|log.Lshortfile),
	}
	driverInitDone := make(chan error)
	go driver.Init(cfg, driverInitDone)
	err := <-driverInitDone
	if err != nil {
		mainlogger.Fatalf("[ERROR] Failed to initalize driver: %v", err)
	}
	mainlogger.Println("[INFO] Driver successfully initialized")
	// Initalize globalstate
	ip, _ := peerdiscovery.GetLocalIP()
	globalstateConfig := globalstate.Config{
		RaftPort: raftPort,
		OwnIP:    ip,
		Floors:   4,
		// OnPromotion:        func() { fmt.Println("PROMOTED!:)") },
		// OnDemotion:         func() { fmt.Println("DEMOTED, :(") },
		OnAquiredConsensus: func() { fmt.Println("Aquired RAFT-consensus") },
		OnLostConsensus:    func() { fmt.Println("Lost RAFT-consensus") },
		OnIncomingCommand:  onIncomingCommand,
		CostFunction:       statetools.CostFunction,
		Logger:             log.New(os.Stderr, "[globalstate] ", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	// Attempt to connect to any known peers.
	if len(peers) > 0 {
		for _, anyPeer := range peers {
			globalstateConfig.InitalPeer = anyPeer.IP + ":" + anyPeer.RaftPort
			mainlogger.Printf("[INFO] Other peers known. Attempting to connect to %s\n", globalstateConfig.InitalPeer)
			break
		}
	}
	gs = globalstate.FSM{}
	err = gs.Init(globalstateConfig)
	if err != nil {
		mainlogger.Printf("[ERROR] Failed to initalize globalstore: %s", err.Error())
	}

	// Start syncinc button leds with the global state.
	go syncBtnLEDs(gs)

	// Block forever
	select {}
}

func syncBtnLEDs(globalstate globalstate.FSM) {
	for {
		time.Sleep(500 * time.Millisecond)
		s, err := globalstate.GetState()
		if err != nil {
			continue
		}
		for k, v := range s.HallUpButtons {
			f, _ := strconv.Atoi(k)
			if v.LastStatus == "done" {
				driver.BtnLEDClear(driver.Btn{Floor: f, Type: driver.HallUp})
			} else {
				driver.BtnLEDSet(driver.Btn{Floor: f, Type: driver.HallUp})
			}
		}

		for k, v := range s.HallDownButtons {
			f, _ := strconv.Atoi(k)
			if v.LastStatus == "done" {
				driver.BtnLEDClear(driver.Btn{Floor: f, Type: driver.HallDown})
			} else {
				driver.BtnLEDSet(driver.Btn{Floor: f, Type: driver.HallDown})
			}
		}

	}
}
