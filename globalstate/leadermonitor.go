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

func (rw *raftwrapper) LeaderMonitor() {
	scanInterval := 500 * time.Millisecond
	orderTimeout := 7 * time.Second
	leaderCh := rw.raft.LeaderCh()
	isLeader := rw.raft.State() == raft.Leader

	// Check inital role and invoke corresponding callback
	if isLeader {
		rw.config.OnPromotion()
	} else {
		rw.config.OnDemotion()
	}

	for {
		if !isLeader {
			// Blocks until assuming leadership
			isLeader = <-leaderCh
			rw.config.OnPromotion()
		}

		// Wait for either loosing leadership or a interval
		select {
		case l := <-leaderCh:
			isLeader = l
			// Call status change callbacks
			if !isLeader {
				rw.config.OnDemotion()
			}

		case <-time.After(scanInterval):
		}

		// Retrieve a working copy of the state
		rw.mu.Lock()
		state := rw.state.DeepCopy()
		rw.mu.Unlock()

		// Inspect unnassigned or orders that have timed out
		unassignedBtns := getUnassignedOrders(state)
		expiredBtns := getTimedOutOrders(state, orderTimeout)

		for _, b := range unassignedBtns {
			lowestCostPeer := rw.config.CostFunction(state, b.Floor, b.Dir)
			sendCmd(b, lowestCostPeer)
		}
		for _, b := range expiredBtns {
			lowestCostPeer := rw.config.CostFunction(state, b.Floor, b.Dir)
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
