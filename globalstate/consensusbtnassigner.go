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
	"time"

	"github.com/hashicorp/raft"
)

func (rw *raftwrapper) ConsensusOrderAssigner(updateBtnStatus func(bs ButtonStatusUpdate) error) {
	// Set intervals. Timeout determined linearly based on number of floors.
	scanInterval := 500 * time.Millisecond
	orderTimeout := time.Duration(3*rw.config.Floors) * time.Second

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
		case <-rw.shutdown:
			return
		}

		// Retrieve a working copy of the state
		rw.mu.Lock()
		state := rw.state.DeepCopy()
		rw.mu.Unlock()

		// Inspect unnassigned or orders that have timed out
		unassignedBtns := getUnassignedOrders(state)
		expiredBtns := getTimedOutOrders(state, orderTimeout)
		var assignees []string

		// Assign to lifts based on cost
		for _, b := range expiredBtns {
			lowestCostPeer := rw.config.CostFunction(state, b.Floor, b.Dir)
			if lowestCostPeer == "" || stringInSlice(lowestCostPeer, assignees) {
				// rw.logger.Printf("[WARN] No lifts currently available to handle order: {Floor:%d, Dir:%s}\n", b.Floor, b.Dir)
				continue
			}
			updateToAssigned(b, lowestCostPeer, updateBtnStatus)
			assignees = append(assignees, lowestCostPeer)
			time.Sleep(100 * time.Millisecond)
			sendCmd(b, lowestCostPeer)

		}
		for _, b := range unassignedBtns {
			lowestCostPeer := rw.config.CostFunction(state, b.Floor, b.Dir)
			if lowestCostPeer == "" || stringInSlice(lowestCostPeer, assignees) {
				// rw.logger.Printf("[WARN] No lifts currently available to handle order: {Floor:%d, Dir:%s}\n", b.Floor, b.Dir)
				continue
			}
			updateToAssigned(b, lowestCostPeer, updateBtnStatus)
			assignees = append(assignees, lowestCostPeer)
			time.Sleep(100 * time.Millisecond)
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
	// Infer address from id (communication endpoint in port above raft-port)
	if strings.Contains(dstNode, ":") == false {
		return fmt.Errorf("bad destination node")
	}
	parts := strings.Split(dstNode, ":")
	raftPort, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("bad destination node: %s", err.Error())
	}
	addr := fmt.Sprintf("%s:%d", parts[0], raftPort+1)

	// Marhsal to json
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(b)
	res, err := http.Post(fmt.Sprintf("http://%s/cmd", addr), "application/json; charset=utf-8", buf)
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	return nil
}

func updateToAssigned(b btn,
	dstNode string,
	updateBtnStatus func(bs ButtonStatusUpdate) error) error {
	bsu := ButtonStatusUpdate{
		Floor:      uint(b.Floor),
		Dir:        b.Dir,
		Status:     BtnStateAssigned,
		AssignedTo: dstNode,
	}
	return updateBtnStatus(bsu)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
