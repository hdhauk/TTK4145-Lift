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

var theFSM *fsm

// Init initalizes the pakcage and stores one instance of the FSM globally in
// the package. Trying to use the globalstate without haveing initialized will
// imideatly rais an error.
func Init(cfg Config) error {
	// Parse port
	rPort := cfg.RaftPort
	rPortStr := strconv.Itoa(rPort)
	cPort := rPort + 1
	cPortStr := strconv.Itoa(cPort)

	// Set up FSM
	theFSM = newFSM(rPortStr)
	theFSM.ownID = cfg.OwnIP + ":" + rPortStr
	theFSM.logger = cfg.Logger
	theFSM.config = cfg

	// Set up storage for FSM
	tmpDir, err1 := ioutil.TempDir("", "raft-fsm-store")
	if err1 != nil {
		theFSM.logger.Printf("[ERROR] Unable to create temporary folder for raft: %v\n", err1.Error())
		return fmt.Errorf("failed to instansiate temp folder: %v", err1)
	}
	defer os.RemoveAll(tmpDir)
	theFSM.RaftDir = tmpDir

	// Start FSM
	if err2 := theFSM.Start(cfg.InitalPeer == ""); err2 != nil {
		theFSM.logger.Printf("[ERROR] Unable to start FSM: %v\n", err2.Error())
		return err2
	}

	// Start the communication service, to handle join requests.
	service := newCommService("127.0.0.1:"+cPortStr, theFSM)
	if err3 := service.Start(); err3 != nil {
		theFSM.logger.Printf("[ERROR] Unable to start communication service: %v\n", err3.Error())
		return err3
	}
	theFSM.logger.Printf("[INFO] Communication service started on on port %v\n", cPort)

	// Join any known peers
	if cfg.InitalPeer != "" {
		err := join(cfg.InitalPeer, rPortStr, cfg.OwnIP, theFSM.logger)
		if err != nil {
			theFSM.logger.Printf("[ERROR] Unable to join node at %s: %s\n", cfg.InitalPeer, err.Error())
			return err
		}
	}
	// Wait for raft to either join or create a new raft. This usually takes 2-3 seconds
	time.Sleep(4 * time.Second)

	// Start the leader worker
	go theFSM.LeaderMonitor()

	theFSM.initDone = true
	return nil
}

func join(initialPeer, raftAddr, ownIP string, logger *log.Logger) error {
	// Marshal join request
	b, err := json.Marshal(map[string]string{"addr": raftAddr})
	if err != nil {
		return err
	}

	// Infer communication port from RaftPort (comport is always one above!)
	parts := strings.Split(initialPeer, ":")
	port, _ := strconv.Atoi(parts[1])
	initialPeer = fmt.Sprintf("%s:%d", parts[0], port+1)

	// HACK: For some reason go struggles to make request to localhost if you
	// try to connect to the actual interface address.
	if strings.Contains(initialPeer, ownIP) {
		parts := strings.Split(initialPeer, ":")
		initialPeer = "127.0.0.1:" + parts[1]
	}

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
		defer resp2.Body.Close()
		return err
	}
	resp2.Body.Close()
	logger.Printf("[INFO] Successfully joined raft\n")
	return nil
}
