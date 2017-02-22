package globalstate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// FSM hold all logic and essentially IS the globalstate handle. You may
// instansiate several FSM as long as they have different ports.
type FSM struct {
	wrapper  *raftwrapper
	comm     *commService
	logger   *log.Logger
	initDone bool
}

// Init sets up and start the FSM.
func (f *FSM) Init(config Config) error {
	// Parse ports
	rPort := config.RaftPort
	rPortStr := strconv.Itoa(rPort)
	cPort := rPort + 1
	cPortStr := strconv.Itoa(cPort)

	// Creating new FSM
	f.wrapper = newRaftWrapper(rPortStr, config.Floors)

	// Safely store the config
	if err := validateConfig(&config); err != nil {
		return err
	}
	f.wrapper.config = config

	// Set basic properties of the fsm
	f.wrapper.ownID = config.OwnIP + ":" + rPortStr
	f.wrapper.logger = config.Logger
	f.logger = config.Logger

	// Set up storage for FSM
	tmpDir, err1 := ioutil.TempDir("", "raft-fsm-store")
	if err1 != nil {
		f.wrapper.logger.Printf("[ERROR] Unable to create temporary folder for raft: %v\n", err1.Error())
		return fmt.Errorf("failed to instansiate temp folder: %v", err1)
	}
	defer os.RemoveAll(tmpDir)
	f.wrapper.RaftDir = tmpDir

	// Start the FSM
	if err := f.wrapper.Start(config.InitalPeer == ""); err != nil {
		f.logger.Printf("[ERROR] Unable to start FSM: %v\n", err.Error())
		return err
	}

	// Start the communication service, to handle join requests.
	f.comm = newCommService("0.0.0.0:"+cPortStr, f.wrapper)
	if err := f.comm.Start(); err != nil {
		f.logger.Printf("[ERROR] Unable to start communication service: %v\n", err.Error())
		return err
	}
	f.logger.Printf("[INFO] Communication service started on on port %v\n", cPort)

	// Join supplied peer.
	if config.InitalPeer != "" {
		err := joinPeerToRaft(config.InitalPeer, rPortStr, config.OwnIP, f.logger)
		if err != nil {
			f.logger.Printf("[ERROR] Unable to join node at %s: %s\n", config.InitalPeer, err.Error())
			return err
		}
	}

	// Wait for raft to either join or create a new raft. This usually takes 2-3 seconds
	time.Sleep(4 * time.Second)

	// Start workers
	go f.wrapper.ConsensusOrderAssigner(f.UpdateButtonStatus)
	go f.wrapper.ConsensusMonitor()

	f.initDone = true
	return nil
}

// Shutdown shuts down the FSM in a safe manner.
func (f *FSM) Shutdown() {
	close(f.wrapper.shutdown)
	f.logger.Println("[INFO] Shutting down raft")
	future := f.wrapper.raft.Shutdown()
	if future.Error() != nil {
		f.logger.Fatalf("[ERROR] Failed to close FSM: %v", future.Error())
	}
	//f.comm.Close()
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

func joinPeerToRaft(initialPeer, raftAddr, ownIP string, logger *log.Logger) error {
	// Marshal join request
	b, err := json.Marshal(map[string]string{"addr": raftAddr})
	if err != nil {
		return err
	}

	// Infer communication port from RaftPort (comport is always one above!)
	parts := strings.Split(initialPeer, ":")
	port, _ := strconv.Atoi(parts[1])
	initialPeer = fmt.Sprintf("%s:%d", parts[0], port+1)

	url := fmt.Sprintf("http://%s/join", initialPeer)
	logger.Printf("[INFO] Attempting to join %v", url)
	resp, err := http.Post(url, "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}

	// Peer admitted on first try (ie. luckily tried the leader on first try)
	if resp.Header.Get("X-Raft-Leader") == "" {
		logger.Printf("[INFO] Successfully joined raft\n")
		return nil
	}

	// Extract address to the leader
	leaderAddr := resp.Header.Get("X-Raft-Leader")
	resp.Body.Close()

	// Request the leader
	url = fmt.Sprintf("http://%s/join", leaderAddr)
	logger.Printf("[INFO] Redirected! Attempting to join: %v\n", url)
	resp2, err := http.Post(url, "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp2.Body.Close()
	logger.Printf("[INFO] Successfully joined raft\n")
	return nil
}
