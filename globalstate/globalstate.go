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

// Update type constants
const (
	// AtFloor is the type for broadcasting new own positon
	AtFloor = iota
	// NewDir is the type for broadcasting a new elevator direction.
	NewDir
	// NewDst is the type for broadcasting a new target destination.
	NewDst
	// NewBtnPrees is the type for broadcasting that a button has been pressed.
	NewBtnPress
	// BtnPressExpedited is the type for broadcasting that an order has been expedited.
	BtnPressExpedited
)

// LiftStatusUpdate defines an message with which you intend to update the global store with.
type LiftStatusUpdate struct {
	Floor uint
	Dst   uint
	Dir   string
}

// GetState returns a copy of the current cluster state.
func GetState() interface{} {
	return theFSM.GetState()
}

// SendLiftStatusUpdate asdasdas asdasd
func SendLiftStatusUpdate(ls LiftStatusUpdate) error {
	// Convert to liftStatus
	status := liftStatus{
		ID:          ownID,
		LastFloor:   ls.Floor,
		Destination: ls.Dst,
		Direction:   ls.Dir,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(status)

	// Get a fresh leader ;)
	<-leaderCh
	leader := leaderComEndpoint(<-leaderCh)
	url := fmt.Sprintf("http://%s/update/lift", leader)
	res, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)

	if res.Header.Get("X-Raft-Leader") == "" {
		return nil
	}
	// TODO: What happens if gets redirected?
	return nil
}

func leaderComEndpoint(leader string) string {
	parts := strings.Split(leader, ":")
	portStr, _ := strconv.Atoi(parts[1])
	return fmt.Sprintf("%s:%d", parts[0], portStr+1)
}
