package globalstate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Init starts the whole shebang!
func Init(anyPeer, joinPort, raftPort string) {

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
	h := NewJoinEndpoint(joinPort, s)
	log.Printf("Starting join endpoint")
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	if anyPeer != "" {
		if err := join(anyPeer, raftPort); err != nil {
			log.Fatalf("failed to join node at %s: %s", anyPeer, err.Error())
		}

	}
}

func join(joinAddr, raftAddr string) error {
	log.Printf("attempting to join")
	b, err := json.Marshal(map[string]string{"addr": raftAddr})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
