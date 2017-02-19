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
	// BtnStateAssigned is a pressed button that the leader have dispatched an lifts to.
	BtnStateAssigned = "assigned"
	// BtnStateDone is a button that is ready to be pressed (ie. no LED lit)
	BtnStateDone = "done"
)

// Config contain all configuration details and callbacks for the FSM.
type Config struct {
	// Port for raft RCP communication. The port above will also be binded, and is used for user-initiated communication over HTTP.
	RaftPort int

	// If supplied the raft will attempt to connect to any cluster the supplied peer is connected to.
	// If left blank the FSM instansiate a brand new raft and elect itself leader.
	InitalPeer string

	// OwnIP may be manually be set. If not supplied it will be infered by the package if needed.
	OwnIP string

	// Number of floors on the lifts in the cluster. Only used for calculating timouts.
	Floors int

	// Called once whenever the node win or loose the raft-leadership.
	OnPromotion func()
	OnDemotion  func()

	// Called once whenever the two leader elections in a row fail to elect a leader.
	OnAquiredConsensus func()

	// Called once whenever the consensus is regained and a leader is elected.
	OnLostConsensus func()

	// Called once whenever the leader have assigned an order to the node.
	OnIncomingCommand func(floor int, dir string)

	// Used by the leader to assign orders.
	CostFunction func(s State, floor int, dir string) string

	Logger *log.Logger

	// Raft may produce a considerable amount of logging, espessialy whenever a node
	// fail to respond. Logging from the globalstate package itself is still active.
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

// UpdateLiftStatus updates the globalstate with the provided liftStatus.
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
