package globalstate

import (
	"io/ioutil"
	"log"
	"os"
	"time"
)

var gstore *Store

// Init intializes the raft-node to be prepared to join a raft-cluster of size
// given by nodes. It sets up all nessesary listeners.
func Init() error {
	//cfg := raft.DefaultConfig()
	gstore := New()
	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	gstore.RaftBind = "127.0.0.1:0"
	gstore.RaftDir = tmpDir

	if err := gstore.Open(true); err != nil {
		log.Fatalf("failed to open store: %s", err)
	}

	// Simple way to ensure there is a leader.
	time.Sleep(3 * time.Second)

	if err := gstore.Set("foo", "bar"); err != nil {
		log.Fatalf("failed to set key: %s", err.Error())
	}

	select {}

}
