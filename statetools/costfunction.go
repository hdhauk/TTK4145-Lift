package statetools

import (
	"math"
	"time"

	"github.com/hdhauk/TTK4145-Lift/globalstate"
)

// CostFunction calculates the best elevator to handle a given order based on the provided state.
func CostFunction(s globalstate.State, floor int, dir string) string {
	lifts := s.Nodes
	costs := make(map[string]int)

	// Calculate cost for all lift
	for _, lift := range lifts {
		costs[lift.ID] = calculateCost(lift, uint(floor), dir, s)
	}

	// Extract lift with lowest cost
	bestLift := ""
	bestCost := 1000
	for liftID, cost := range costs {
		if cost < bestCost {
			bestCost = cost
			bestLift = liftID
		}
	}

	// Make sure no other buttons are assigned to this liftID
	if hasOtherAssignments(s, bestLift) {
		return ""
	} else if bestCost > 99 {
		return ""
	}

	return bestLift
}

func calculateCost(
	lift globalstate.LiftStatus,
	floor uint, dir string,
	state globalstate.State) int {

	// Start with zero cost and penalize as we proceed
	cost := 0

	// Have the been alive recently?
	if time.Since(lift.LastUpdate) > time.Second*10 {
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
