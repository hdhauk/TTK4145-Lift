package globalstate

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_InitSingleFSM(t *testing.T) {
	consensus := make(chan bool)
	config1 := Config{
		RaftPort:           9000,
		OnAquiredConsensus: func() { consensus <- true },
		OnLostConsensus:    func() { consensus <- false },
		CostFunction:       func(s State, f int, d string) string { return "localhost:8003" },
		Logger:             log.New(ioutil.Discard, "[globalstate1] ", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft1 := FSM{}
	if err := raft1.Init(config1); err != nil {
		t.Errorf("failed to initalize FSM")
	}

	res := <-consensus
	if !res {
		t.Errorf("failed to obtain consensus as single node")
	}
}

func Test_TwoNodeSimpleConsensus(t *testing.T) {
	consensus := make(chan bool, 2)
	config1 := Config{
		RaftPort:           9002,
		OnAquiredConsensus: func() { consensus <- true },
		OnLostConsensus:    func() { consensus <- false },
		CostFunction:       func(s State, f int, d string) string { return "localhost:8003" },
		Logger:             log.New(ioutil.Discard, "[globalstate1] ", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft1 := FSM{}
	raft1.Init(config1)

	time.Sleep(7 * time.Second)
	config2 := Config{
		RaftPort:           9004,
		InitalPeer:         getOutboundIP() + ":9002",
		OnAquiredConsensus: func() { consensus <- true },
		OnLostConsensus:    func() { consensus <- false },
		CostFunction:       func(s State, f int, d string) string { return "localhost:8005" },
		Logger:             log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft2 := FSM{}
	raft2.Init(config2)
	i := 0
	for {
		c := <-consensus
		if !c {
			t.Errorf("failed to obtain consensus")
		}
		i++
		if i == 2 {
			return
		}
	}
}

func Test_TwoNodeConsensusWithTraffic(t *testing.T) {
	config1 := Config{
		RaftPort:           9006,
		CostFunction:       func(s State, f int, d string) string { return "localhost:8003" },
		Logger:             log.New(ioutil.Discard, "[globalstate1] ", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft1 := FSM{}
	raft1.Init(config1)

	time.Sleep(5 * time.Second)
	config2 := Config{
		RaftPort:           9008,
		InitalPeer:         getOutboundIP() + ":9006",
		CostFunction:       func(s State, f int, d string) string { return "localhost:8005" },
		Logger:             log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft2 := FSM{}
	raft2.Init(config2)

	raft1.UpdateButtonStatus(ButtonStatusUpdate{2, "up", "done"})
	raft2.UpdateButtonStatus(ButtonStatusUpdate{1, "down", "assigned"})
	raft1.UpdateLiftStatus(LiftStatusUpdate{1, "stop", 2, ""})
	raft2.UpdateLiftStatus(LiftStatusUpdate{3, "down", 1, "up"})

	time.Sleep(1 * time.Second)
	state1, _ := raft1.GetState()
	state2, _ := raft2.GetState()
	assert.Equal(t, state1, state2)
}

func Test_ThreeNodeClusterWithRedirect(t *testing.T) {
	config1 := Config{
		RaftPort:           9010,
		CostFunction:       func(s State, f int, d string) string { return "localhost:8003" },
		Logger:             log.New(ioutil.Discard, "[globalstate1] ", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft1 := FSM{}
	raft1.Init(config1)

	time.Sleep(5 * time.Second)
	config2 := Config{
		RaftPort:           9012,
		InitalPeer:         getOutboundIP() + ":9010",
		CostFunction:       func(s State, f int, d string) string { return "localhost:8005" },
		Logger:             log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft2 := FSM{}
	raft2.Init(config2)

	time.Sleep(5 * time.Second)
	config3 := Config{
		RaftPort:           9014,
		InitalPeer:         getOutboundIP() + ":9012",
		CostFunction:       func(s State, f int, d string) string { return "localhost:8005" },
		Logger:             log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		DisableRaftLogging: true,
	}
	raft3 := FSM{}
	raft3.Init(config3)

	blankState, _ := raft1.GetState()
	raft1.UpdateButtonStatus(ButtonStatusUpdate{2, "up", "done"})
	time.Sleep(1 * time.Second)
	state1, _ := raft1.GetState()
	state2, _ := raft2.GetState()
	state3, _ := raft3.GetState()
	assert.NotEqual(t, blankState, state1)
	assert.Equal(t, state1, state2)
	assert.Equal(t, state2, state3)

}
