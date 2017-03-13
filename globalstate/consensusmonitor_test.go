package globalstate

import (
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func Test_LoosingConsensus(t *testing.T) {
	consensusLost := make(chan interface{})
	config1 := Config{
		RaftPort:           9016,
		OnLostConsensus:    func() { consensusLost <- true },
		CostFunction:       func(s State, f int, d string) string { return "localhost:8003" },
		Logger:             log.New(ioutil.Discard, "[globalstate1] ", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft1 := FSM{}
	raft1.Init(config1)

	// Initialize two new nodes, and unwind their scope to kill them afterwards.

	time.Sleep(3 * time.Second)
	config2 := Config{
		RaftPort:           9018,
		InitalPeer:         getOutboundIP() + ":9016",
		CostFunction:       func(s State, f int, d string) string { return "localhost:8005" },
		Logger:             log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft2 := FSM{}
	raft2.Init(config2)

	time.Sleep(3 * time.Second)
	config3 := Config{
		RaftPort:           9020,
		InitalPeer:         getOutboundIP() + ":9016",
		CostFunction:       func(s State, f int, d string) string { return "localhost:8005" },
		Logger:             log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft3 := FSM{}
	raft3.Init(config3)
	time.Sleep(3 * time.Second)
	raft2.Shutdown()
	raft3.Shutdown()

	select {
	case <-consensusLost:
		return
	case <-time.After(5 * time.Second):
		t.Fatalf("Did not detect consensus loss within 2 seconds")
	}

}
