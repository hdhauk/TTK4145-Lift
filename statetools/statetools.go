package statetools

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/hdhauk/TTK4145-Lift/globalstate"
)

const (
	down = "down"
	up   = "up"
	done = "done"
)

// LocalState is a thread safe data structure almost identical to the global state.
type LocalState struct {
	state globalstate.State
	mu    sync.Mutex
}

// NewLocalState instantiate a new local state ready for use.
func NewLocalState() *LocalState {
	l := LocalState{}
	l.state.HallUpButtons = make(map[string]globalstate.Status)
	l.state.HallDownButtons = make(map[string]globalstate.Status)
	return &l
}

// UpdateButtonStatus is analogous to its counterpart in the globalstate package.
func (ls *LocalState) UpdateButtonStatus(bsu globalstate.ButtonStatusUpdate) error {
	if ls == nil {
		return fmt.Errorf("state not instantiated")
	}
	floor := strconv.Itoa(int(bsu.Floor))

	ls.mu.Lock()
	switch bsu.Dir {
	case up:
		ls.state.HallUpButtons[floor] = globalstate.Status{
			LastChange: time.Now(),
			LastStatus: bsu.Status,
		}
	case down:
		ls.state.HallDownButtons[floor] = globalstate.Status{
			LastChange: time.Now(),
			LastStatus: bsu.Status,
		}
	}
	ls.mu.Unlock()
	return nil
}

// GetNextOrder returns the oldest of any buttons on the state that are not in the
// state "done".
func (ls *LocalState) GetNextOrder() (floor int, dir string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	oldest := time.Now()
	floor = -1
	dir = ""
	for floorStr, s := range ls.state.HallUpButtons {
		if s.LastStatus == globalstate.BtnStateUnassigned && s.LastChange.Before(oldest) {
			oldest = s.LastChange
			floor, _ = strconv.Atoi(floorStr)
			dir = up
		}
	}
	for floorStr, s := range ls.state.HallDownButtons {
		if s.LastStatus == globalstate.BtnStateUnassigned && s.LastChange.Before(oldest) {
			oldest = s.LastChange
			floor, _ = strconv.Atoi(floorStr)
			dir = down
		}
	}
	return
}

// GetShareworthyUpdates dumps all orders in the local state that aren't currently marked
// as done. For convenience they are returned as an array of button status updates.
func (ls *LocalState) GetShareworthyUpdates() []globalstate.ButtonStatusUpdate {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	var share = []globalstate.ButtonStatusUpdate{}

	for floorStr, s := range ls.state.HallUpButtons {
		if s.LastStatus != globalstate.BtnStateDone {
			floor, _ := strconv.Atoi(floorStr)
			bsu := globalstate.ButtonStatusUpdate{
				Floor:  uint(floor),
				Dir:    up,
				Status: globalstate.BtnStateUnassigned,
			}
			share = append(share, bsu)
		}
	}
	for floorStr, s := range ls.state.HallDownButtons {
		if s.LastStatus != globalstate.BtnStateDone {
			floor, _ := strconv.Atoi(floorStr)
			bsu := globalstate.ButtonStatusUpdate{
				Floor:  uint(floor),
				Dir:    down,
				Status: globalstate.BtnStateUnassigned,
			}
			share = append(share, bsu)
		}
	}
	return share
}

// CloneGlobalstate clone the entire provided globalstate onto the local state such
// that the local state is now an exact copy of the global one.
func (ls *LocalState) CloneGlobalstate(gs globalstate.State) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.state = gs

	// Mark all assigned orders as unassigned
	for floor, s := range ls.state.HallDownButtons {
		if s.LastStatus == globalstate.BtnStateAssigned {
			ls.state.HallDownButtons[floor] = globalstate.Status{LastStatus: globalstate.BtnStateUnassigned, LastChange: time.Now()}
		}
	}
	for floor, s := range ls.state.HallUpButtons {
		if s.LastStatus == globalstate.BtnStateAssigned {
			ls.state.HallUpButtons[floor] = globalstate.Status{LastStatus: globalstate.BtnStateUnassigned, LastChange: time.Now()}
		}
	}
}
