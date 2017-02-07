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
)

var leaderCh = make(chan string)
var ownID string // TODO: move into fsm...
var theFSM *fsm

// Init TODO: description
func Init(cfg Config) error {
	// Parse port
	rPort := cfg.RaftPort
	rPortStr := strconv.Itoa(rPort)
	cPort := rPort + 1
	cPortStr := strconv.Itoa(cPort)

	ownID = cfg.OwnIP + ":" + rPortStr

	// Set up FSM
	theFSM = newFSM(rPortStr)

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

	go func() {
		for {
			leaderCh <- theFSM.GetLeader()
		}
	}()

	// TODO: Implement worker here....
	select {}

}

func join(joinAddr, raftAddr, ownIP string, logger *log.Logger) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr})
	if err != nil {
		return err
	}

	// HACK: For some reason go struggles to make request to localhost if you
	// try to connect to the actual interface address.
	if strings.Contains(joinAddr, ownIP) {
		parts := strings.Split(joinAddr, ":")
		joinAddr = "127.0.0.1:" + parts[1]
	}

	url := fmt.Sprintf("http://%s/join", joinAddr)
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
