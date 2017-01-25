// Package globalstate is wrapper package for CoreOS' implementation of the
// Raft consensus protocol. See https://github.com/hashicorp/raft
// For a description of how the Raft algorithm works see:
//  - http://thesecretlivesofdata.com/raft/
//  - https://raft.github.io/
//  - https://raft.github.io/raft.pdf
// TL;DR:
//  Raft provide an algorithm for ensuring consensus in the cluser, which we in
//  this project use for keeping track of:
//    - Last registered floor for all elevators
//    - Whether an elevator is at a standstill or moving somewhere
//    - What buttons are pressed in each floor.
package globalstate

import "time"

// Init intializes the raft-node to be prepared to join a raft-cluster of size
// given by nodes. It sets up all nessesary listeners.
func Init(nodes int) {
	// cfg := raft.DefaultConfig()

}

// CommitToState emits the given Update to the current cluster leader. It will
// return an error of the leader is unreachable, or if it fail to recieve an
//acknowledgement that the Update is committed to the cluster.
func CommitToState(u Update) error {
	return nil
}

// GetState returns a copy of the current cluster state.
func GetState() State {
	return State{}
}

// Update defines all messages that may be sendt to the cluster.
type Update struct {
	// Type may be: "FLOOR", "MOTOR", "ORDER"
	Type string
}

// State defines the centralized state managed by the raft-cluster
type State struct {
	// Number of floors for all elevators
	Floors uint
	// ClusterSize is the number of nodes in the cluster
	ClusterSize uint
	// Nodes is the IP:port of all nodes in the system
	Nodes []elevator
	// HallUpButtons, true of they are lit. Equivalent with an order there
	HallUpButtons  []status
	HallDownButton []status
}

type elevator struct {
	id            string
	lastFloor     uint
	lastDirection uint
}

type status struct {
	assignedTo string    // elevator.id
	lastStatus string    // "UNASSIGNED", "ASSIGNED", "DONE"
	lastChange time.Time //
}
