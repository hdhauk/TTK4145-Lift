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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Public facing data types and constants
// =============================================================================

const (
	// BtnStateUnassigned is an button that have been pressed but no further action taken.
	BtnStateUnassigned = "unassigned"
	// BtnStateAssigned is a pressed button that the leader have dispatched an elevator to.
	BtnStateAssigned = "assigned"
	// BtnStateDone is a button that is ready to be pressed (ie. no LED lit)
	BtnStateDone = "done"
)

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

// LiftStatusUpdate defines an message with which you intend to update the global store with.
type LiftStatusUpdate struct {
	Floor uint
	Dst   uint
	Dir   string
}

// ButtonStatusUpdate defines a message with which you intend to update the global store with.
type ButtonStatusUpdate struct {
	Floor  uint
	Dir    string
	Status string
}

// Public facing functions
// =============================================================================

// UpdateLiftStatus asdasdas asdasd
func UpdateLiftStatus(ls LiftStatusUpdate) error {
	if !theFSM.initDone {
		return fmt.Errorf("globalstate not yet initalized")
	}
	// Convert to liftStatus
	status := LiftStatus{
		ID:          theFSM.ownID,
		LastFloor:   ls.Floor,
		Destination: ls.Dst,
		Direction:   ls.Dir,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(status)

	leader := leaderComEndpoint(theFSM.GetLeader())
	url := fmt.Sprintf("http://%s/update/lift", leader)
	res, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	return nil
}

// UpdateButtonStatus update the globalt store  with the supplied button update.
// If unable to reach the raft-leader it will return an error.
func UpdateButtonStatus(bs ButtonStatusUpdate) error {
	if !theFSM.initDone {
		return fmt.Errorf("globalstate not yet initalized")
	}
	// Marshal for sending as json
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(bs)
	if err != nil {
		theFSM.logger.Printf("[ERROR] Unable to marshal button status update: %s\n", err.Error())
		return err
	}

	// Post the update to the current raft-leader
	leader := leaderComEndpoint(theFSM.GetLeader())
	url := fmt.Sprintf("http://%s/update/button", leader)
	res, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		theFSM.logger.Printf("[ERROR] Unable to send button status update to leader: %s\n", err.Error())
		return err
	}
	io.Copy(os.Stdout, res.Body)
	return nil
}

// GetState returns a copy of the current cluster state.
func GetState() (State, error) {
	if !theFSM.initDone {
		return State{}, fmt.Errorf("globalstate not yet initalized")
	}
	return theFSM.GetState(), nil
}

// Helper functions
// =============================================================================
func leaderComEndpoint(leader string) string {
	parts := strings.Split(leader, ":")
	portStr, _ := strconv.Atoi(parts[1])
	return fmt.Sprintf("%s:%d", parts[0], portStr+1)
}
