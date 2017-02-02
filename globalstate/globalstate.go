/*
Package globalstate is wrapper package for Hashicorps' implementation of the
Raft consensus protocol. See https://github.com/hashicorp/raft.

For a description of how the Raft algorithm works see:
 - http://thesecretlivesofdata.com/raft/
 - https://raft.github.io/
 - https://raft.github.io/raft.pdf

TL;DR:

	Raft provide an algorithm for ensuring consensus in the cluser, which we in
	this project use for keeping track of:
	* Last registered floor for all elevators
	* Whether an elevator is at a standstill or moving somewhere
	* What buttons are pressed in each floor.
*/
package globalstate

import (
	"fmt"
	"time"
)

// Add emits the given Update to the current cluster leader. It will
// return an error of the leader is unreachable, or if it fail to recieve an
//acknowledgement that the Update is committed to the cluster.
func Add() error {
	fmt.Println("Attempting to add a key-value pair...")
	if err := gstore.Set("testKey", "testValue"); err != nil {
		return err
	}
	return nil
}

// GetState returns a copy of the current cluster state.
func GetState() interface{} {
	//return gstore.Snapshot()
	return nil
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
	Nodes []Elevator
	// HallUpButtons, true of they are lit. Equivalent with an order there
	HallUpButtons  []Status
	HallDownButton []Status
}

// Elevator defines the publicly available information about the elevators in the cluster.
type Elevator struct {
	ID            string
	LastFloor     uint
	LastDirection uint
}

// Status defines the status of a button.
//All buttons of the same type on the same floor are considered equal,
//and as long as the elevator is online will behave the exact same way.
// ie. will pressing the up-button at floor 3 on one elevator yield the same
// result as pressing the same button on another elevator.
type Status struct {
	AssignedTo string    // elevator.id
	LastStatus string    // "UNASSIGNED", "ASSIGNED", "DONE"
	LastChange time.Time //
}

// DispatchOrder dispatches an order to the provided elevator.
// The function will do nothing if the elevator isn't the master.
func DispatchOrder(floor int, dir string, elevatorID string) error {
	return nil
}
