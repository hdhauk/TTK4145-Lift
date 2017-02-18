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
	"log"
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
	RaftPort           int
	InitalPeer         string
	OwnIP              string
	Floors             int
	OnPromotion        func()
	OnDemotion         func()
	OnAquiredConsensus func()
	OnLostConsensus    func()
	OnIncomingCommand  func(floor int, dir string)
	CostFunction       func(s State, floor int, dir string) string
	Logger             *log.Logger
	DisableRaftLogging bool
}

// LiftStatusUpdate defines an message with which you intend to update the global store with.
type LiftStatusUpdate struct {
	CurrentFloor uint
	CurrentDir   string
	DstFloor     uint
	DstBtnDir    string
}

// ButtonStatusUpdate defines a message with which you intend to update the global store with.
type ButtonStatusUpdate struct {
	Floor      uint
	Dir        string
	Status     string
	AssignedTo string
}

// Public facing functions
// =============================================================================

// UpdateLiftStatus asdasdas asdasd
func (f *FSM) UpdateLiftStatus(ls LiftStatusUpdate) error {
	if !f.initDone {
		return fmt.Errorf("globalstate not yet initalized")
	}
	// Convert to liftStatus
	status := LiftStatus{
		ID:                         f.wrapper.ownID,
		LastFloor:                  ls.CurrentFloor,
		Direction:                  ls.CurrentDir,
		DestinationFloor:           ls.DstFloor,
		DestinationButtonDirection: ls.DstBtnDir,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(status)

	// Get leader communication endpoint
	leader, err := f.wrapper.leaderComEndpoint()
	if err != nil {
		return err
	}

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
func (f *FSM) UpdateButtonStatus(bs ButtonStatusUpdate) error {
	if !f.initDone {
		return fmt.Errorf("globalstate not yet initalized")
	}
	// Marshal for sending as json
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(bs)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to marshal button status update: %s\n", err.Error())
		return err
	}

	// Get leader communication endpoint
	leader, err := f.wrapper.leaderComEndpoint()
	if err != nil {
		return err
	}

	// Post the update to the current raft-leader
	url := fmt.Sprintf("http://%s/update/button", leader)
	res, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		f.logger.Printf("[ERROR] Unable to send button status update to leader: %s\n", err.Error())
		return err
	}
	io.Copy(os.Stdout, res.Body)
	return nil
}

// GetState returns a copy of the current cluster state.
func (f *FSM) GetState() (State, error) {
	if !f.initDone {
		return State{}, fmt.Errorf("globalstate not yet initalized")
	}
	return f.wrapper.GetState(), nil
}

// Helper functions
// =============================================================================
func (rw *raftwrapper) leaderComEndpoint() (string, error) {
	leaderRaftAddr := rw.GetLeader()
	if leaderRaftAddr == "" {
		return "", fmt.Errorf("Cannot send button status. No current leader")
	}
	parts := strings.Split(leaderRaftAddr, ":")
	portStr, _ := strconv.Atoi(parts[1])
	return fmt.Sprintf("%s:%d", parts[0], portStr+1), nil
}
