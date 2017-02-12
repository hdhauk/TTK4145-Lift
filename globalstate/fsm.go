package globalstate

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

// fsm is the datastructure that hold pretty much everything interesting in
// this package. The actual data is stored in the `state`-variable and read/writes
// to it is protected by the mutex. The interface with the raft-library is DONE
// through the raft-object, and itself provides replication hand heartbeating.
// fsm need to fulfill the raft.fsm interface:
//    * fsm.Appy(*raftLog) interface{}
//    * fsm.Snapshot()(FSMSnapshot, error)
//    * fsm.Restore(io.ReadCloser) error
type fsm struct {
	RaftDir  string
	RaftPort string
	mu       sync.Mutex
	state    State
	raft     *raft.Raft
	logger   *log.Logger
	initDone bool
	ownID    string
	config   Config
}

// newFSM return a new raft-enabled finite state machine.
func newFSM(rPortStr string) *fsm {
	s := State{
		Floors:          4,
		Nodes:           make(map[string]LiftStatus),
		HallUpButtons:   make(map[string]Status),
		HallDownButtons: make(map[string]Status),
	}
	return &fsm{
		RaftPort: rPortStr,
		state:    s,
		logger:   log.New(os.Stderr, "[globalstate] ", log.Ltime|log.Lshortfile),
	}
}

func (f *fsm) Start(enableSingle bool) error {
	// Set up Raft configuration
	raftCfg := raft.DefaultConfig()
	raftCfg.Logger = f.logger

	// Set up Raft communication.
	rSocket := ":" + f.RaftPort
	addr, err := net.ResolveTCPAddr("tcp", rSocket)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to resolve TCP raft-endpoint: %s\n", err.Error())
		return err
	}
	trans, err := raft.NewTCPTransportWithLogger(rSocket, addr, 3, 5*time.Second, f.logger)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to set up Raft TCP Transport: %v\n", err.Error())
		return err
	}

	// Create peer storage
	peerStore := raft.NewJSONPeers(f.RaftDir, trans)

	// Enable single-mode in order to allow bootstrapping of a new raft-cluster
	// if no other peers are provided during initialization.
	if enableSingle {
		f.logger.Println("[INFO] Starting raft with single-node mode enabled.")
		raftCfg.EnableSingleNode = true
		raftCfg.DisableBootstrapAfterElect = false
	}

	// Create a snapshot store, allowing the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStoreWithLogger(f.RaftDir, 2, f.logger)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to create Snapshot store: %v\n", err.Error())
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create log store and stable store. These only exist in memory, as our
	// database is useless for old information anyway.
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Instantiate Raft
	ra, err := raft.NewRaft(raftCfg, f, logStore, stableStore, snapshots, peerStore, trans)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to instansiate raft: %v\n", err.Error())
		return fmt.Errorf("new raft: %s", err)
	}
	f.raft = ra
	f.logger.Println("[INFO] Successfully initalized Raft")
	return nil
}

// raft-interface functions
// =============================================================================

// Apply applies a Raft log entry to the key-value store.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		f.logger.Fatalf(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Type {
	case "updateFloor":
		return f.applyUpdateFloor(c.Value)
	case "nodeUpdate":
		return f.applyNodeUpdate(c.Key, c.Value)
	case "btnUpUpdate":
		return f.applyBtnUpUpdate(c.Key, c.Value)
	case "btnDownUpdate":
		return f.applyBtnDownUpdate(c.Key, c.Value)
	default:
		f.logger.Printf(fmt.Sprintf("Unrecognized command: %s", c.Type))
		return nil
	}
}

// Snapshot is used to support log compaction. This call should
// return an FSMSnapshot which can be used to save a point-in-time
// snapshot of the FSM. Apply and Snapshot are not called in multiple
// threads, but Apply will be called concurrently with Persist. This means
// the FSM should be implemented in a fashion that allows for concurrent
// updates while a snapshot is happening.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return &fsmSnapshot{store: f.state.DeepCopy()}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	newState := State{}
	if err := json.NewDecoder(rc).Decode(&newState); err != nil {
		f.logger.Printf("Failed to decode FSM from snapshot: %v\n", err)
		return err
	}
	// No need to lock the mutex as this command isn't run concurrently with any
	// other command (according to Hashicorp docs)
	f.state = newState
	return nil
}

// Functions for general usage and update of the raft-fsm
// =============================================================================

// Getstatus TODO: description
func (f *fsm) GetStatus() uint32 {
	return uint32(f.raft.State())
}

// GetLeader TODO: description
func (f *fsm) GetLeader() string {
	return f.raft.Leader()
}

// Join joins a node, located at addr, to this store. The node must be ready to
// respond to Raft communications at that address.
func (f *fsm) Join(addr string) error {
	f.logger.Printf("[INFO] Recieved join request from remote node %s\n", addr)
	future := f.raft.AddPeer(addr)
	if future.Error() != nil {
		f.logger.Printf("[WARN] Unable to add peer: %v\n", future.Error())
		return future.Error()
	}
	f.logger.Printf("[INFO] Successfully joined node %s to the raft.\n", addr)
	return nil
}

func (f *fsm) GetState() State {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.state.DeepCopy()
}

func (f *fsm) UpdateLiftStatus(ls LiftStatus) error {
	// Make sure the node currently hold leadership.
	if f.raft.State() != raft.Leader {
		f.logger.Printf("[WARN] Unable to update lift status. Not currently leader.\n")
		return fmt.Errorf("not leader")
	}

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
		f.logger.Printf("[ERROR] Failed to encode json: %s\n", err.Error())
		return err
	}

	// Apply command to raft
	future := f.raft.Apply(b, 5*time.Second)
	return future.Error()
}

func (f *fsm) UpdateButtonStatus(bsu ButtonStatusUpdate) error {
	// Make sure the node currently hold leadership.
	if f.raft.State() != raft.Leader {
		f.logger.Printf("[WARN] Unable to update button status. Not currently leader.\n")
		return fmt.Errorf("not leader")
	}

	// Determine Direction
	var t string
	if bsu.Dir == "up" || bsu.Dir == "UP" {
		t = "btnUpUpdate"
	} else if bsu.Dir == "down" || bsu.Dir == "DOWN" {
		t = "btnDownUpdate"
	} else {
		f.logger.Printf("[ERROR] Unable to parse direction in button update: %s\n", bsu.Dir)
		return fmt.Errorf("unable to parse direction: %s", bsu.Dir)
	}

	// Create status
	status := Status{
		AssignedTo: "",
		LastStatus: bsu.Status,
		LastChange: time.Now(),
	}

	// Marshal payload to bytes. No need for errorcheck, as it it just recently
	// been unmarhaled by the same library.
	v, err := json.Marshal(status)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to marshal status: %s\n", err.Error())
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
		f.logger.Printf("[ERROR] Failed to marshal raft log command: %s\n", err.Error())
		return err
	}

	// Apply command to raft
	future := f.raft.Apply(b, 5*time.Second)
	return future.Error()

}

// Internal fsm-function
// 	These are functions called by the raft apply command in order to recreate
//	the store based on the raft-log. Functins here are called by the
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

func (f *fsm) applyUpdateFloor(floor interface{}) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	floorInt, _ := floor.(uint)
	// TODO: Implement test for the floorInt
	// if !ok {
	// 	f.logger.Printf("[ERROR] Unable to apply floorUpdate. Bad floor: %v\n", floor)
	// 	return nil
	// }
	f.state.Floors = floorInt
	return nil
}

func (f *fsm) applyNodeUpdate(nodeID string, e []byte) interface{} {
	// Unmarshal liftstatus
	var lift LiftStatus
	err := json.Unmarshal(e, &lift)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to unmarhal liftStatus: %s\n", err.Error())
		return fmt.Errorf("unable to unmarshal liftstats")
	}

	// Update the actual datastore entry
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Nodes[nodeID] = lift
	return nil
}

func (f *fsm) applyBtnUpUpdate(floor string, b []byte) interface{} {
	// Unmarshal ButtonStatusUpdate
	var status Status
	err := json.Unmarshal(b, &status)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to unmarshal status: %s\n", err.Error())
		return fmt.Errorf("unable to unmarshal ButtonUpStatusUpdate: %s", err.Error())
	}

	// Update the actual datastore entry
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.HallUpButtons[floor] = status
	return nil
}

func (f *fsm) applyBtnDownUpdate(floor string, b []byte) interface{} {
	// Unmarshal ButtonStatusUpdate
	var status Status
	err := json.Unmarshal(b, &status)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to unmarshal status: %s\n", err.Error())
		return fmt.Errorf("unable to unmarshal ButtonDownStatusUpdate: %s", err.Error())
	}

	// Update the actual datastore entry
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.HallDownButtons[floor] = status
	return nil
}
