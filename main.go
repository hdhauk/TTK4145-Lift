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

// Global state is threadsafe and for simplicity available to the whole main package.
var gs globalstate.FSM
var ls *statetools.LocalState
var mainlogger = log.New(os.Stderr, "[main] ", log.Ltime|log.Lshortfile)
var goToCh = make(chan driver.Btn, 9)
var goToFromInsideCh = make(chan driver.Btn, 9)
var orderDoneCh = make(chan interface{})
var haveConsensusBtnSyncCh = make(chan bool)
var haveConsensusAssignerCh = make(chan bool)

func main() {
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
	driverConfig := driver.Config{
		Floors:       9,
		OnBtnPress:   onBtnPress,
		OnNewStatus:  onNewStatus,
		OnDstReached: onDstReached,
		Logger:       log.New(os.Stderr, "[driver] ", log.Ltime|log.Lshortfile),
	}
	if simPort != "" {
		driverConfig.SimMode = true
		driverConfig.SimPort = simPort
	}

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
		Floors:             9,
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
	gs = globalstate.FSM{}
	err = gs.Init(globalstateConfig)
	if err != nil {
		mainlogger.Printf("[ERROR] Failed to initalize globalstore: %s", err.Error())
	}

	// Set up local state in case network connection is lost.
	ls = statetools.NewLocalState()

	// Start syncinc button leds with the global state.
	go syncBtnLEDs(gs)
	go liftDriver()
	go noConsensusAssigner()

	// Block forever
	select {}
}

func syncBtnLEDs(globalstate globalstate.FSM) {
	online := false
	for {
		select {
		case b := <-haveConsensusBtnSyncCh:
			online = b
		case <-time.After(1 * time.Microsecond):

		}
		if !online {
			continue
		}

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

func liftDriver() {
	outsideQueue := btnQueue{}
	insideQueue := btnQueue{}
	ready := true

	for {
		select {
		case dst := <-goToCh:
			outsideQueue.Queue(dst)
		case dst := <-goToFromInsideCh:
			insideQueue.Queue(dst)
		case <-orderDoneCh:
			ready = true
		}
		if dst, empty := insideQueue.Dequeue(); !empty && ready {
			driver.GoToFloor(dst.Floor, "")
			ready = false
		} else if dst, empty := outsideQueue.Dequeue(); !empty && ready {
			driver.GoToFloor(dst.Floor, dst.Type.String())
			ready = false
		}
	}
}

func noConsensusAssigner() {
	online := true
	for {

		select {
		case b := <-haveConsensusAssignerCh:
			online = b
		case <-time.After(1 * time.Millisecond):
		}
		if online {
			continue
		}

		floor, dir := ls.GetNextOrder()
		if dir == "up" {
			goToCh <- driver.Btn{Floor: floor, Type: driver.HallUp}
		} else if dir == "down" {
			goToCh <- driver.Btn{Floor: floor, Type: driver.HallDown}
		}

	}
}

type btnQueue struct {
	btns []driver.Btn
}

func (bq *btnQueue) Queue(b driver.Btn) {
	bq.btns = append(bq.btns, b)
}
func (bq *btnQueue) Dequeue() (b driver.Btn, empty bool) {
	if len(bq.btns) == 0 {
		return driver.Btn{}, true
	}
	b = bq.btns[0]
	bq.btns = bq.btns[1:]
	return
}
