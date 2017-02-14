package globalstate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

// StandAloneFSM asdasd
type StandAloneFSM struct {
	f *fsm
	s *service
	l *log.Logger
}

// Init deras
func (saf *StandAloneFSM) Init(config Config) error {
	// Parse ports
	rPort := config.RaftPort
	rPortStr := strconv.Itoa(rPort)
	cPort := rPort + 1
	cPortStr := strconv.Itoa(cPort)

	// Creating new FSM
	saf.f = newFSM(rPortStr)

	// Safely store the config
	if err := validateConfig(&config); err != nil {
		return err
	}
	fmt.Println(config)
	saf.f.config = config

	// Set basic properties of the fsm
	saf.f.ownID = config.OwnIP + ":" + rPortStr
	saf.f.logger = config.Logger
	saf.l = config.Logger

	// Set up storage for FSM
	tmpDir, err1 := ioutil.TempDir("", "raft-fsm-store")
	if err1 != nil {
		saf.f.logger.Printf("[ERROR] Unable to create temporary folder for raft: %v\n", err1.Error())
		return fmt.Errorf("failed to instansiate temp folder: %v", err1)
	}
	defer os.RemoveAll(tmpDir)
	saf.f.RaftDir = tmpDir

	// Start the FSM
	if err := saf.f.Start(config.InitalPeer == ""); err != nil {
		theFSM.logger.Printf("[ERROR] Unable to start FSM: %v\n", err.Error())
		return err
	}

	// Start the communication service, to handle join requests.
	saf.s = newCommService("0.0.0.0:"+cPortStr, saf.f)
	if err := saf.s.Start(); err != nil {
		saf.l.Printf("[ERROR] Unable to start communication service: %v\n", err.Error())
		return err
	}
	saf.l.Printf("[INFO] Communication service started on on port %v\n", cPort)

	// Wait for raft to either join or create a new raft. This usually takes 2-3 seconds
	time.Sleep(4 * time.Second)

	// Start workers
	go saf.f.LeaderMonitor()
	go saf.f.ConsensusMonitor()

	saf.f.initDone = true
	return nil

}

func validateConfig(c *Config) error {
	if c.RaftPort == 0 {
		return fmt.Errorf("no raft port set")
	}
	if c.OwnIP == "" {
		c.OwnIP = getOutboundIP()
	}
	if c.OnPromotion == nil {
		c.OnPromotion = func() {}
	}
	if c.OnDemotion == nil {
		c.OnDemotion = func() {}
	}
	if c.OnAquiredConsensus == nil {
		c.OnAquiredConsensus = func() {}
	}
	if c.OnLostConsensus == nil {
		c.OnLostConsensus = func() {}
	}
	if c.OnIncomingCommand == nil {
		c.OnIncomingCommand = func(f int, d string) {}
	}
	if c.CostFunction == nil {
		c.CostFunction = func(s State, f int, d string) string { return "localhost:8000" }
	}
	if c.Logger == nil {
		c.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return nil
}
