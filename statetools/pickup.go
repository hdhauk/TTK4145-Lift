package statetools

import (
	"strconv"

	"github.com/hdhauk/TTK4145-Lift/globalstate"
)

// ShouldStopAndPickup returns true if there are an opportunity to pick up someone in
// the provided floor.
func ShouldStopAndPickup(s globalstate.State, currentFloor int, currentDir string) bool {
	// Extract applicable buttons
	var buttons map[string]globalstate.Status

	if currentDir == "UP" || currentDir == "up" {
		buttons = s.HallUpButtons
	} else if currentDir == "DOWN" || currentDir == "down" {
		buttons = s.HallDownButtons
	} else {
		return false
	}

	// Get status for the current floor
	status, ok := buttons[strconv.Itoa(currentFloor)]
	if !ok {
		return false
	}
	if status.LastStatus == "done" {
		return false
	}
	return true
}
