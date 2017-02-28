package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
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
var floors int

// Pick ports randomly
var raftPort = 1024 + rand.Intn(64510)

// Both the global and local state are threadsafe and for convenience thus
// available to the whole main package.
var stateGlobal globalstate.FSM
var stateLocal *statetools.LocalState

// Set up looging. All packages have their own logger with prefix: [package name]
var mainlogger = log.New(os.Stderr, "[main] ", log.Ltime|log.Lshortfile)

// Set up internal communication in package main.
// All communication with other packages are done through callbacks.
var goToCh = make(chan driver.Btn, 9)
var orderDoneCh = make(chan interface{})
var haveConsensusBtnSyncCh = make(chan bool)
var haveConsensusAssignerCh = make(chan bool)

func main() {
	// Parse command line argument flags
	flag.StringVar(&nick, "nick", strconv.Itoa(os.Getpid()), "Nickname of this peer. Default is the process id (PID)")
	flag.StringVar(&simPort, "sim", "", "Listening port of the simulator")
	flag.IntVar(&raftPort, "raft", raftPort, "Communication port for raft")
	flag.IntVar(&floors, "floors", 4, "Number of floors on the lift.")
	flag.Parse()
	mainlogger.Printf("[INFO] Raft port: %d, Nickname: %s, Simulator port: %s, Floors: %d\n", raftPort, nick, simPort, floors)

	// Initialize peer discovery. Any discovered are only used for initializing
	// the global store.
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
	time.Sleep(2 * discoveryConfig.BroadcastInterval) // Allow for detection of any remote peers

	// Initialize driver
	driverConfig := driver.Config{
		Floors:       floors,
		OnBtnPress:   onBtnPress,
		OnNewStatus:  onNewStatus,
		OnDstReached: onDstReached,
		Logger:       log.New(os.Stderr, "[driver] ", log.Ltime|log.Lshortfile),
	}
	if simPort != "" {
		driverConfig.SimMode = true
		driverConfig.SimPort = simPort
	}
	// Start driver and wait for it to complete initialization.
	driverInitDone := make(chan error)
	go driver.Init(driverConfig, driverInitDone)
	err := <-driverInitDone
	if err != nil {
		mainlogger.Fatalf("[ERROR] Failed to initalize driver: %v", err)
	}
	mainlogger.Println("[INFO] Driver successfully initialized")

	// Initalize globalstate
	ip, _ := peerdiscovery.GetLocalIP()
	globalstateConfig := globalstate.Config{
		RaftPort:           raftPort,
		OwnIP:              ip,
		Floors:             floors,
		OnAquiredConsensus: onAquiredConsensus,
		OnLostConsensus:    onLostConsensus,
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
	stateGlobal = globalstate.FSM{}
	err = stateGlobal.Init(globalstateConfig)
	if err != nil {
		mainlogger.Printf("[ERROR] Failed to initalize globalstore: %s", err.Error())
	}

	// Set up local state in case network connection is lost.
	stateLocal = statetools.NewLocalState()

	// Start workers for coordination
	go syncBtnLEDs(stateGlobal) // Only active when consensus is achieved.
	go orderQueuer()            // Always active.
	go noConsensusAssigner()    // Only active when consensus is missing.

	// Capture Ctrl+C in order to stop the lift if it is moving.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		driver.Stop()
		mainlogger.Fatalf("[WARN] Interrupt detected. Stopping lift and exiting.\n")
	}()

	// Block forever
	select {}
}
