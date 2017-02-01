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

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

var globalRaft raft.Raft

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// Config defines what
type Config struct {
	ClusterSize int
	OnPromotion func()
	OnDemotion  func()
}

type store struct {
	RaftDir  string
	RaftPort string

	mu sync.Mutex
	// FIXME
	m map[string]string // The actual store

	raft *raft.Raft

	logger *log.Logger
}

// Creates a new store
func new() *store {
	return &store{
		// FIXME
		m:      make(map[string]string),
		logger: log.New(os.Stderr, "[store] ", log.LstdFlags),
	}
}

func (s *store) Open(enableSingle bool) error {
	// Setup Raft configuration
	cfg := raft.DefaultConfig()

	// Setup Raft Communication
	addr, err := net.ResolveTCPAddr("tcp", s.RaftPort)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(s.RaftPort, addr, 2, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create peer storage
	peerStore := raft.NewJSONPeers(s.RaftDir, transport)

	// Check for any existing peerStore
	peers, err := peerStore.Peers()
	if err != nil {
		return err
	}

	// Allow the node to entry single-mode, potentially electing itself, if
	// explicitly enabled and there is only 1 node in the cluster already.
	if enableSingle && len(peers) <= 1 {
		s.logger.Println("enabling single-node mode")
		cfg.EnableSingleNode = true
		cfg.DisableBootstrapAfterElect = false
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot stroe %s", err)
	}

	// Create the log store and stable store.
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(s.RaftDir, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}

	// Instantiate the Raft systems
	ra, err := raft.NewRaft(cfg, (*fsm)(s), logStore, logStore, snapshots, peerStore, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra

	return nil
}

func (s *store) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[key], nil
}

func (s *store) Set(key, value string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}

	c := &command{
		Op:    "set",
		Key:   key,
		Value: value,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

// NOTE: probably not useful...
func (s *store) Delete(key string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}

	c := &command{
		Op:  "delete",
		Key: key,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

// Join joins a node, located at addr, to this store. The node must be ready to
// respond to Raft communications at that address.
func (s *store) Join(addr string) error {
	s.logger.Printf("recieved join request for remote node at %s", addr)

	f := s.raft.AddPeer(addr)
	if f.Error() != nil {
		return f.Error()
	}
	s.logger.Printf("node at %s joined successfully", addr)
	return nil
}

type fsm store

// apply applies a Raft log entry to the key-value store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		// FIXME: Seems excessive to panic here...
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "set":
		return f.applySet(c.Key, c.Value)
	case "delete":
		return f.applyDelete(c.Key)
	default:
		// FIXME: Seems excessive to panic here...
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}
}

// snapshot returns a snapshot of the key-value store
// TODO: No way this can return an error
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Lock()

	// Clone the map.
	// FIXME
	o := make(map[string]string)
	for k, v := range f.m {
		o[k] = v
	}
	return &fsmSnapshot{store: o}, nil
}

// restore roll the key-value store back to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	// FIXME
	o := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required accoring to Hashicorp docs...
	f.m = o
	return nil
}

// TODO: why the return value?
func (f *fsm) applySet(key, value string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, key)
	return nil
}

// TODO: Why the return value?
func (f *fsm) applyDelete(key string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, key)
	return nil
}

type fsmSnapshot struct {
	// FIXME
	store map[string]string
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode Data
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		if err := sink.Close(); err != nil {
			return err
		}
		return nil
	}()

	if err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (f *fsmSnapshot) Release() {}
