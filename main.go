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

	// Block forever
	select {}
}
