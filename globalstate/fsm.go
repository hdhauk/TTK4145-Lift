package globalstate

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// fsm is the datastructure that hold pretty much everything interesting in
// this package.
// it need to fulfill the raft.fsm interface:
//    * fsm.Appy(*raftLog) interface{}        // TODO: Testing
//    * fsm.Snapshot()(FSMSnapshot, error)    // TODO: Testing
//    * fsm.Restore(io.ReadCloser) error      // TODO: Testing
type fsm struct {
	RaftDir  string
	RaftPort string
	mu       sync.Mutex
	state    State
	raft     *raft.Raft
	logger   *log.Logger
}

func newFSM(rPortStr string) *fsm {
	return &fsm{
		RaftPort: rPortStr,
		logger:   log.New(os.Stderr, "globalstore -->", log.Ltime|log.Lshortfile),
	}
}

func (f *fsm) Start(enableSingle bool) error {
	// Set up Raft configuration
	raftCfg := raft.DefaultConfig()

	// Set up Raft communication.
	rSocket := "127.0.0.1:" + f.RaftPort
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

	// Create log store and stable store
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(f.RaftDir, "raft.db"))
	if err != nil {
		f.logger.Printf("[ERROR] Unable to create boltDB log and stable store: %v\n", err.Error())
		return fmt.Errorf("new bolt store: %s", err)
	}

	// Instantiate Raft
	ra, err := raft.NewRaft(raftCfg, f, logStore, logStore, snapshots, peerStore, trans)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to instansiate raft: %v\n", err.Error())
		return fmt.Errorf("new raft: %s", err)
	}
	f.raft = ra
	f.logger.Println("[INFO] Successfully initalized Raft")
	return nil
}

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
		return f.appyBtnDownUpdate(c.Key, c.Value)
	default:
		f.logger.Printf(fmt.Sprintf("Unrecognized command: %s", c.Type))
		return nil
	}
}

// command defines actions that the raft.Log contain, and a series of
// commands should be able to recreate the State.
type command struct {
	/*
	  Types:
	  - "updateFloor": key=<Don't Care>   Value=<floors int>
	  - "nodeUpdate":  key=<up:raftport>  Value=<struct{ID string, LastFloor, Destination uint}>
	  - "btnUpUpdate": key=<floor>        Value=<struct{AssignedTo, LastStatus string, LastChange time.Time}>
	  - "btnDownUpdate": key=<floor>      Value=<struct{AssignedTo, LastStatus string, LastChange time.Time}>
	*/
	Type  string      `json:"type,omitempty"`
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

func (f *fsm) applyUpdateFloor(floor interface{}) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	floorInt, ok := floor.(uint)
	if !ok {
		f.logger.Printf("[ERROR] Unable to apply floorUpdate. Bad floor: %v\n", floor)
		return nil
	}
	f.state.Floors = floorInt
	return nil
}

func (f *fsm) applyNodeUpdate(nodeID string, e interface{}) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	elev, ok := e.(Elevator)
	if !ok {
		f.logger.Printf("[ERROR] Unable to apply nodeUpdate. Bad node: nodeID=%v , Elevator=%v", nodeID, e)
		return nil
	}
	f.state.Nodes[nodeID] = elev.DeepCopy()
	return nil
}

func (f *fsm) applyBtnUpUpdate(floor string, s interface{}) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	btnStatus, ok := s.(Status)
	if !ok {
		f.logger.Printf("[ERROR] Unable to apply btnUpUpdate. Bad status object: floor=%s , status=%v", floor, s)
		return nil
	}
	f.state.HallUpButtons[floor] = btnStatus.DeepCopy()
	return nil
}

func (f *fsm) appyBtnDownUpdate(floor string, s interface{}) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	btnStatus, ok := s.(Status)
	if !ok {
		f.logger.Printf("[ERROR] Unable to apply btnDownUpdate. Bad status object: floor=%s , status=%v", floor, s)
		return nil
	}
	f.state.HallDownButton[floor] = btnStatus.DeepCopy()
	return nil
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
