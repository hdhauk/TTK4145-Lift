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

// Config defines ...TODO: Something informative here...
type Config struct {
	RaftPort        int
	InitalPeer      string
	OwnIP           string
	OnPromotion     func()
	OnDemotion      func()
	IncomingCommand func(floor int)
	CostFunction    func(State) string
}

// Add emits the given Update to the current cluster leader. It will
// return an error of the leader is unreachable, or if it fail to recieve an
//acknowledgement that the Update is committed to the cluster.
func Add() error {
	// fmt.Println("Attempting to add a key-value pair...")
	// if err := gstore.Set("testKey", "testValue"); err != nil {
	// 	return err
	// }
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

// DispatchOrder dispatches an order to the provided elevator.
// The function will do nothing if the elevator isn't the master.
func DispatchOrder(floor int, dir string, elevatorID string) error {
	return nil
}
