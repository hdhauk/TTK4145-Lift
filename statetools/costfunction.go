package statetools

import (
	"fmt"
	"math"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
)

// CostFunction sdasda
func CostFunction(s globalstate.State, floor int, dir string) string {
	lifts := s.Nodes
	costs := make(map[string]int)

	// Caluclate cost for all lift
	for _, lift := range lifts {
		costs[lift.ID] = calculateCost(lift, uint(floor), dir, s)
	}

	// Extract lift with lowest cost
	bestLift := ""
	bestCost := 99
	for liftID, cost := range costs {
		if cost < bestCost {
			bestLift = liftID
			bestCost = cost
		}
	}

	// Make sure no other buttons are assigned to this liftID
	if bestLift != "" && hasOtherAssignments(s, bestLift) {
		fmt.Printf("Best lift (%s) have other assignments already\n", bestLift)
		return ""
	}

	return bestLift
}

func calculateCost(
	lift globalstate.LiftStatus,
	floor uint, dir string,
	state globalstate.State) int {

	// Start with zero cost and penalize
	cost := 0

	// Have the been alive recently?
	if time.Since(lift.LastUpdate) > time.Second*6 {
		return 100
	}

	// Is the lift busy with another order?
	if lift.DestinationButtonDirection != "" {
		return 105
	}

	// How far is the lift from the button?
	cost += int(math.Abs(float64(lift.LastFloor) - float64(floor)))

	return cost
}

func hasOtherAssignments(state globalstate.State, liftID string) bool {
	for _, v := range state.HallUpButtons {
		if v.AssignedTo == liftID {
			return true
		}
	}
	for _, v := range state.HallDownButtons {
		if v.AssignedTo == liftID {
			return true
		}
	}
	return false
}
