package statetools

import "time"

// State asdasd
type State struct {
	// Number of floors for all elevators
	Floors uint
	// Nodes is the IP:port of all nodes in the system
	Nodes map[string]LiftStatus
	// HallUpButtons, true of they are lit. Equivalent with an order there
	HallUpButtons   map[string]Status
	HallDownButtons map[string]Status
}

// Status asdasd
type Status struct {
	AssignedTo string    // elevator.id
	LastStatus string    // "unassigned", "assigned", "done"
	LastChange time.Time //
}

// LiftStatus sadsad
type LiftStatus struct {
	ID                         string
	LastFloor                  uint
	Direction                  string
	DestinationFloor           uint
	DestinationButtonDirection string
	LastUpdate                 time.Time
}

// GetAssignment calulates the elevator with the lowest cost for a the provided state and order.
func GetAssignment(s State, floor int, dir string) string {
	return ""
}

// AddOrder adds an order to the provided state. NB: Only for use on local states.
func AddOrder(s State, floor int, dir string) {

}

// RemoveOrder removes an order from the provided state. NB: Only for use on local states.
func RemoveOrder(s State, floor int, dir string) {

}
