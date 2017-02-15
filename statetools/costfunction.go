package statetools

import "time"

// CostFunction sdasda
func CostFunction(s State, floor int, dir string) string {
	lifts := s.Nodes
	costs := make(map[string]int)

	// Caluclate cost for all lift
	for _, lift := range lifts {
		costs[lift.ID] = calculateCost(lift, uint(floor), dir)
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

	return bestLift
}

func calculateCost(lift LiftStatus, floor uint, dir string) int {
	cost := 0

	// Have the been alive recently?
	if time.Since(lift.LastUpdate) > time.Second*6 {
		return 100
	}

	// Is the lift busy with another order?
	if lift.DestinationButtonDirection != "" {
		return 100
	}

	// Is the lift moving?
	if lift.Direction != "stop" {
		cost++
	}

	// How far is the lift from the button?
	cost += int(float64(lift.LastFloor) - float64(floor))

	return cost
}
