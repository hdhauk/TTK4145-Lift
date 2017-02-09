package globalstate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/raft"
)

func (f *fsm) LeaderMonitor() {
	scanInterval := 500 * time.Millisecond
	orderTimeout := 7 * time.Second
	leaderCh := f.raft.LeaderCh()
	isLeader := f.raft.State() == raft.Leader

	// Check inital role and invoke corresponding callback
	if isLeader {
		f.config.OnPromotion()
	} else {
		f.config.OnDemotion()
	}

	for {
		if !isLeader {
			// Blocks until assuming leadership
			isLeader = <-leaderCh
			f.config.OnPromotion()
		}

		// Wait for either loosing leadership or a interval
		select {
		case l := <-leaderCh:
			isLeader = l
			// Call status change callbacks
			if !isLeader {
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

		for _, b := range unassignedBtns {
			lowestCostPeer := f.config.CostFunction(state, b.Floor, b.Dir)
			sendCmd(b, lowestCostPeer)
		}
		for _, b := range expiredBtns {
			lowestCostPeer := f.config.CostFunction(state, b.Floor, b.Dir)
			sendCmd(b, lowestCostPeer)
		}

	}
}

type btn struct {
	Floor int
	Dir   string
}

func getUnassignedOrders(s State) []btn {
	var btns []btn
	// Scan down buttons
	for k, v := range s.HallDownButtons {
		if v.LastStatus == BtnStateUnassigned {
			f, _ := strconv.Atoi(k)
			btns = append(btns, btn{Floor: f, Dir: "down"})
		}
	}

	// Scan up buttons
	for k, v := range s.HallUpButtons {
		if v.LastStatus == BtnStateUnassigned {
			f, _ := strconv.Atoi(k)
			btns = append(btns, btn{Floor: f, Dir: "up"})
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
			btns = append(btns, btn{Floor: f, Dir: "down"})
		}
	}

	// Scan up buttons
	for k, v := range s.HallUpButtons {
		if v.LastStatus == BtnStateAssigned &&
			time.Since(v.LastChange) > timeout {
			f, _ := strconv.Atoi(k)
			btns = append(btns, btn{Floor: f, Dir: "up"})
		}
	}
	return btns
}

func sendCmd(b btn, dstNode string) error {
	// Marhsal to json
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(b)
	res, err := http.Post(fmt.Sprintf("http://%s/cmd", dstNode), "application/json; charset=utf-8", buf)
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	return nil
}
