package globalstate

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/raft"
)

func (f *fsm) LeaderMonitor() {
	scanInterval := 500 * time.Millisecond
	orderTimeout := 7 * time.Second
	leaderCh := f.raft.LeaderCh()
	isLeader := f.raft.State() == raft.Leader
	for {
		if !isLeader {
			// Blocks until assuming leadership
			isLeader = <-leaderCh
		}

		// Wait for either loosing leadership or a interval
		select {
		case l := <-leaderCh:
			isLeader = l
			// Call status change callbacks
			if isLeader {
				f.config.OnPromotion()
			} else {
				f.config.OnDemotion()
			}

		case <-time.After(scanInterval):
		}

		// Retrieve a working copy of the state
		f.mu.Lock()
		state := f.state.DeepCopy()
		f.mu.Unlock()

		// Inspect unnassigned or orders that have timed out
		unassignedBtns := getUnassignedOrders(state)
		expiredBtns := getTimedOutOrders(state, orderTimeout)

		fmt.Printf("Unnassigned buttons: %+v\n", unassignedBtns)
		fmt.Printf("Expired buttons: %+v\n", expiredBtns)

	}
}

type btn struct {
	floor int
	dir   string
}

func getUnassignedOrders(s State) []btn {
	var btns []btn
	// Scan down buttons
	for k, v := range s.HallDownButtons {
		if v.LastStatus == BtnStateUnassigned {
			f, _ := strconv.Atoi(k)
			btns = append(btns, btn{floor: f, dir: "down"})
		}
	}

	// Scan up buttons
	for k, v := range s.HallUpButtons {
		if v.LastStatus == BtnStateUnassigned {
			f, _ := strconv.Atoi(k)
			btns = append(btns, btn{floor: f, dir: "up"})
		}
	}

	return btns
}

func getTimedOutOrders(s State, timeout time.Duration) []btn {
	var btns []btn

	// Scan down buttons
	for k, v := range s.HallDownButtons {
		if v.LastStatus == BtnStateAssigned &&
			time.Since(v.LastChange) > timeout {
			f, _ := strconv.Atoi(k)
			btns = append(btns, btn{floor: f, dir: "down"})
		}
	}

	// Scan up buttons
	for k, v := range s.HallUpButtons {
		if v.LastStatus == BtnStateAssigned &&
			time.Since(v.LastChange) > timeout {
			f, _ := strconv.Atoi(k)
			btns = append(btns, btn{floor: f, dir: "up"})
		}
	}
	return btns
}
