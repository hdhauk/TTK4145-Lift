package globalstate

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

// raftwrapper is the data structure that hold pretty much everything interesting in
// this package. The actual data is stored in the `state`-variable and read/writes
// to it is protected by the mutex. The interface with the raft-library is DONE
// through the raft-object, and itself provides replication and heartbeating.
// raftwrapper need to fulfill the raft.raftwrapper interface:
//    * raftwrapper.Appy(*raftLog) interface{}
//    * raftwrapper.Snapshot()(FSMSnapshot, error)
//    * raftwrapper.Restore(io.ReadCloser) error
type raftwrapper struct {
	RaftDir  string
	RaftPort string
	mu       sync.Mutex
	state    State
	raft     *raft.Raft
	logger   *log.Logger
	ownID    string
	config   Config
	shutdown chan interface{}
}

// newRaftWrapper return a new raft-enabled finite state machine.
func newRaftWrapper(rPortStr string, floors int) *raftwrapper {
	s := State{
		Floors:          uint(floors),
		Nodes:           make(map[string]LiftStatus),
		HallUpButtons:   make(map[string]Status),
		HallDownButtons: make(map[string]Status),
	}
	return &raftwrapper{
		RaftPort: rPortStr,
		state:    s,
		logger:   log.New(os.Stderr, "[globalstate] ", log.Ltime|log.Lshortfile),
		shutdown: make(chan interface{}),
	}
}

func (rw *raftwrapper) Start(enableSingle bool) error {
	// Set up Raft configuration
	raftCfg := raft.DefaultConfig()
	if rw.config.DisableRaftLogging {
		raftCfg.Logger = log.New(ioutil.Discard, "", log.Ltime)
	} else {
		raftCfg.Logger = log.New(os.Stderr, "[raft] ", log.Ltime|log.Lshortfile)
	}

	// Set up Raft communication.
	rSocket := ":" + rw.RaftPort
	localIP := getOutboundIP()
	addr, err := net.ResolveTCPAddr("tcp", localIP+rSocket)
	if err != nil {
		rw.logger.Printf("[ERROR] Unable to resolve TCP raft-endpoint: %s\n", err.Error())
		return err
	}
	trans, err := raft.NewTCPTransportWithLogger(rSocket, addr, 3, 5*time.Second, rw.logger)
	if err != nil {
		rw.logger.Printf("[ERROR] Unable to set up Raft TCP Transport: %v\n", err.Error())
		return err
	}

	// Create peer storage
	peerStore := raft.NewJSONPeers(rw.RaftDir, trans)

	// Enable single-mode in order to allow bootstrapping of a new raft-cluster
	// if no other peers are provided during initialization.
	if enableSingle {
		rw.logger.Println("[INFO] Starting raft with single-node mode enabled.")
		raftCfg.EnableSingleNode = true
		raftCfg.DisableBootstrapAfterElect = false
	}

	// Create a snapshot store, allowing the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStoreWithLogger(rw.RaftDir, 2, rw.logger)
	if err != nil {
		rw.logger.Printf("[ERROR] Unable to create Snapshot store: %v\n", err.Error())
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create log store and stable store. These only exist in memory, as our
	// database is useless for old information anyway.
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Instantiate Raft
	ra, err := raft.NewRaft(raftCfg, rw, logStore, stableStore, snapshots, peerStore, trans)
	if err != nil {
		rw.logger.Printf("[ERROR] Unable to instansiate raft: %v\n", err.Error())
		return fmt.Errorf("new raft: %s", err)
	}
	rw.raft = ra
	rw.logger.Println("[INFO] Successfully initialized Raft")
	return nil
}

// raft-interface functions
// =============================================================================

// Apply applies a Raft log entry to the key-value store.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (rw *raftwrapper) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		rw.logger.Fatalf(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Type {
	case "updateFloor":
		return rw.applyUpdateFloor(c.Value)
	case "nodeUpdate":
		return rw.applyNodeUpdate(c.Key, c.Value)
	case "btnUpUpdate":
		return rw.applyBtnUpUpdate(c.Key, c.Value)
	case "btnDownUpdate":
		return rw.applyBtnDownUpdate(c.Key, c.Value)
	default:
		rw.logger.Printf(fmt.Sprintf("Unrecognized command: %s", c.Type))
		return nil
	}
}

// Snapshot is used to support log compaction. This call should
// return an FSMSnapshot which can be used to save a point-in-time
// snapshot of the FSM. Apply and Snapshot are not called in multiple
// threads, but Apply will be called concurrently with Persist. This means
// the FSM should be implemented in a fashion that allows for concurrent
// updates while a snapshot is happening.
func (rw *raftwrapper) Snapshot() (raft.FSMSnapshot, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return &fsmSnapshot{store: rw.state.DeepCopy()}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (rw *raftwrapper) Restore(rc io.ReadCloser) error {
	newState := State{}
	if err := json.NewDecoder(rc).Decode(&newState); err != nil {
		rw.logger.Printf("Failed to decode FSM from snapshot: %v\n", err)
		return err
	}
	// No need to lock the mutex as this command isn't run concurrently with any
	// other command (according to Hashicorp docs)
	rw.state = newState
	return nil
}

// Functions for general usage and update of the raft-fsm
// =============================================================================

// Getstatus returns the current raft-status (leader, candidate or follower)
func (rw *raftwrapper) GetStatus() uint32 {
	return uint32(rw.raft.State())
}

// GetLeader returns the ip:port of the current leader
func (rw *raftwrapper) GetLeader() string {
	return rw.raft.Leader()
}

// Join joins a node, located at addr, to this store. The node must be ready to
// respond to Raft communications at that address.
func (rw *raftwrapper) Join(addr string) error {
	future := rw.raft.AddPeer(addr)
	if future.Error() != nil {
		rw.logger.Printf("[WARN] Unable to add peer: %v\n", future.Error())
		return future.Error()
	}
	rw.logger.Printf("[INFO] Successfully joined node %s to the raft.\n", addr)
	return nil
}

// GetState returns a copy of the full state as it currently stands.
func (rw *raftwrapper) GetState() State {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.state.DeepCopy()
}

func (rw *raftwrapper) UpdateLiftStatus(ls LiftStatus) error {
	// Make sure the node currently hold leadership.
	if rw.raft.State() != raft.Leader {
		rw.logger.Printf("[WARN] Unable to update lift status. Not currently leader.\n")
		return fmt.Errorf("not leader")
	}

	ls.LastUpdate = time.Now()
	v, _ := json.Marshal(ls)

	// Create log entry for raft.
	cmd := &command{
		Type:  "nodeUpdate",
		Key:   ls.ID,
		Value: v,
	}

	// Encode to json
	b, err := json.Marshal(cmd)
	if err != nil {
		rw.logger.Printf("[ERROR] Failed to encode json: %s\n", err.Error())
		return err
	}

	// Apply command to raft
	future := rw.raft.Apply(b, 5*time.Second)
	return future.Error()
}

func (rw *raftwrapper) UpdateButtonStatus(bsu ButtonStatusUpdate) error {
	// Make sure the node currently hold leadership.
	if rw.raft.State() != raft.Leader {
		rw.logger.Printf("[WARN] Unable to update button status. Not currently leader.\n")
		return fmt.Errorf("not leader")
	}

	// Determine Direction
	var t string
	if bsu.Dir == "up" || bsu.Dir == "UP" {
		t = "btnUpUpdate"
	} else if bsu.Dir == "down" || bsu.Dir == "DOWN" {
		t = "btnDownUpdate"
	} else {
		rw.logger.Printf("[ERROR] Unable to parse direction in button update: %s\n", bsu.Dir)
		return fmt.Errorf("unable to parse direction: %s", bsu.Dir)
	}

	// Create status
	status := Status{
		AssignedTo: bsu.AssignedTo,
		LastStatus: bsu.Status,
		LastChange: time.Now(),
	}

	// Marshal payload to bytes. No need for error check, as it it just recently
	// been unmarhaled by the same library.
	v, err := json.Marshal(status)
	if err != nil {
		rw.logger.Printf("[ERROR] Unable to marshal status: %s\n", err.Error())
		return err
	}

	// Create log entry command for raft.
	cmd := &command{
		Type:  t,
		Key:   strconv.Itoa(int(bsu.Floor)),
		Value: v,
	}

	// Marshal command to json
	b, err := json.Marshal(cmd)
	if err != nil {
		rw.logger.Printf("[ERROR] Failed to marshal raft log command: %s\n", err.Error())
		return err
	}

	// Apply command to raft
	future := rw.raft.Apply(b, 5*time.Second)
	return future.Error()

}

// Internal fsm-function
// 	These are functions called by the raft apply command in order to recreate
//	the store based on the raft-log. Functions here are called by the
//	Apply()-command, and should not be called directly.
// =============================================================================

// command defines actions that the raft.Log contain, and a series of
// commands should be able to recreate the State.
type command struct {
	/*
		Types:
		- "updateFloor": key=<Don't Care>   Value=<floors int>
		- "nodeUpdate":  key=<ip:raftport>  Value=<struct{ID string, LastFloor, Destination uint}>
		- "btnUpUpdate": key=<floor>        Value=<struct{AssignedTo, LastStatus string, LastChange time.Time}>
		- "btnDownUpdate": key=<floor>      Value=<struct{AssignedTo, LastStatus string, LastChange time.Time}>
	*/
	Type  string `json:"type,omitempty"`
	Key   string `json:"key,omitempty"`
	Value []byte `json:"value,omitempty"`
}

func (rw *raftwrapper) applyUpdateFloor(floor interface{}) interface{} {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	floorInt, _ := floor.(uint)
	rw.state.Floors = floorInt
	return nil
}

func (rw *raftwrapper) applyNodeUpdate(nodeID string, e []byte) interface{} {
	// Unmarshal liftstatus
	var lift LiftStatus
	err := json.Unmarshal(e, &lift)
	if err != nil {
		rw.logger.Printf("[ERROR] Unable to unmarhal liftStatus: %s\n", err.Error())
		return fmt.Errorf("unable to unmarshal liftstats")
	}

	// Update the actual data store entry
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.state.Nodes[nodeID] = lift
	return nil
}

func (rw *raftwrapper) applyBtnUpUpdate(floor string, b []byte) interface{} {
	// Unmarshal ButtonStatusUpdate
	var status Status
	err := json.Unmarshal(b, &status)
	if err != nil {
		rw.logger.Printf("[ERROR] Unable to unmarshal status: %s\n", err.Error())
		return fmt.Errorf("unable to unmarshal ButtonUpStatusUpdate: %s", err.Error())
	}

	// Update the actual data store entry
	rw.mu.Lock()
	defer rw.mu.Unlock()
	// Discard any transitions from "done" --> "assigned"
	if rw.state.HallUpButtons[floor].LastStatus == BtnStateDone && status.LastStatus == BtnStateAssigned {
		return nil
	}
	rw.state.HallUpButtons[floor] = status
	return nil
}

func (rw *raftwrapper) applyBtnDownUpdate(floor string, b []byte) interface{} {
	// Unmarshal ButtonStatusUpdate
	var status Status
	err := json.Unmarshal(b, &status)
	if err != nil {
		rw.logger.Printf("[ERROR] Unable to unmarshal status: %s\n", err.Error())
		return fmt.Errorf("unable to unmarshal ButtonDownStatusUpdate: %s", err.Error())
	}

	// Update the actual data store entry
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Discard any transitions from "done" --> "assigned"
	if rw.state.HallDownButtons[floor].LastStatus == BtnStateDone && status.LastStatus == BtnStateAssigned {
		return nil
	}
	rw.state.HallDownButtons[floor] = status
	return nil
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")

	return localAddr[0:idx]
}
