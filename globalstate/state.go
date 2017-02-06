package globalstate

import "time"

// State defines the centralized state managed by the raft-cluster
type State struct {
	// Number of floors for all elevators
	Floors uint
	// Nodes is the IP:port of all nodes in the system
	Nodes map[string]Elevator
	// HallUpButtons, true of they are lit. Equivalent with an order there
	HallUpButtons  map[string]Status
	HallDownButton map[string]Status
}

// DeepCopy safely return a copy of the state
func (s *State) DeepCopy() State {
	var nodes map[string]Elevator
	for k, v := range s.Nodes {
		nodes[k] = v.DeepCopy()
	}
	var hallUp map[string]Status
	for k, v := range s.HallUpButtons {
		hallUp[k] = v.DeepCopy()
	}
	var hallDown map[string]Status
	for k, v := range s.HallDownButton {
		hallDown[k] = v.DeepCopy()
	}
	return State{
		Floors:         s.Floors,
		Nodes:          nodes,
		HallUpButtons:  hallUp,
		HallDownButton: hallDown,
	}
}

// Status defines the status of a button.
//All buttons of the same type on the same floor are considered equal,
//and as long as the elevator is online will behave the exact same way.
// ie. will pressing the up-button at floor 3 on one elevator yield the same
// result as pressing the same button on another elevator.
type Status struct {
	AssignedTo string    // elevator.id
	LastStatus string    // "UNASSIGNED", "ASSIGNED", "DONE"
	LastChange time.Time //
}

// DeepCopy safely return a copy of the Status
func (s *Status) DeepCopy() Status {
	return Status{
		AssignedTo: s.AssignedTo,
		LastStatus: s.LastStatus,
		LastChange: s.LastChange,
	}
}

// Elevator defines the publicly available information about the elevators in the cluster.
type Elevator struct {
	ID          string
	LastFloor   uint
	Destination uint
}

// DeepCopy safely return a copy of the elevator.
func (e *Elevator) DeepCopy() Elevator {
	return Elevator{
		ID:          e.ID,
		LastFloor:   e.LastFloor,
		Destination: e.Destination,
	}
}
