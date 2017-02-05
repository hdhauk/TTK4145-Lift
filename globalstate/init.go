package globalstate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var leaderCh chan bool

// Init starts the whole shebang!
func Init(anyPeer, joinPort, raftPort, ownIP string) {

	// Set up the raft-storage
	s := NewStore()
	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)
	s.RaftBind = "127.0.0.1:" + raftPort
	s.RaftDir = tmpDir
	if err := s.Open(anyPeer == ""); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	// Start http endpoint
	service := newCommService("127.0.0.1:"+joinPort, s)
	s.logger.Println("Starting communication endpoint")
	if err := service.Start(); err != nil {
		s.logger.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	// Join any known peers
	if anyPeer != "" {
		if err := join(anyPeer, raftPort, ownIP); err != nil {
			log.Fatalf("failed to join node at %s: %s", anyPeer, err.Error())
		}
	}

	// Trigger callbacks
	// leaderCh = s.raft.LeaderCh()
	// select{
	// case isLeader := <-leaderCh:
	// 	s.
	// }

}

func join(joinAddr, raftAddr, ownIP string) error {
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
	log.Printf("[INFO] Attempting to join %v", url)
	resp, err := http.Post(url, "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}

	// Peer admitted on first try (ie. luckily tried the leader on first try)
	if resp.Header.Get("X-Raft-Leader") == "" {
		log.Printf("[INFO] Successfully joined raft\n")
		return nil
	}

	// Extract address to the leader
	leaderAddr := resp.Header.Get("X-Raft-Leader")
	resp.Body.Close()

	// Request the leader
	url = fmt.Sprintf("http://%s/join", leaderAddr)
	log.Printf("[INFO] Redirected! Attempting to join: %v\n", url)
	resp2, err := http.Post(url, "application-type/json", bytes.NewReader(b))
	if err != nil {
		defer resp2.Body.Close()
		return err
	}
	resp2.Body.Close()
	log.Printf("[INFO] Successfully joined raft\n")
	return nil
}
