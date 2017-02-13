package globalstate

import "time"

// State defines the centralized state managed by the raft-cluster
type State struct {
	// Number of floors for all elevators
	Floors uint
	// Nodes is the IP:port of all nodes in the system
	Nodes map[string]LiftStatus
	// HallUpButtons, true of they are lit. Equivalent with an order there
	HallUpButtons   map[string]Status
	HallDownButtons map[string]Status
}

// NewState returns a new state
func NewState(floors uint) *State {
	s := State{
		Floors:          floors,
		Nodes:           make(map[string]LiftStatus),
		HallUpButtons:   make(map[string]Status),
		HallDownButtons: make(map[string]Status),
	}
	return &s
}

// DeepCopy safely return a copy of the state
func (s *State) DeepCopy() State {
	nodes := make(map[string]LiftStatus)
	for k, v := range s.Nodes {
		nodes[k] = v.DeepCopy()
	}

	hallUp := make(map[string]Status)
	for k, v := range s.HallUpButtons {
		hallUp[k] = v.DeepCopy()
	}
	hallDown := make(map[string]Status)
	for k, v := range s.HallDownButtons {
		hallDown[k] = v.DeepCopy()
	}
	return State{
		Floors:          s.Floors,
		Nodes:           nodes,
		HallUpButtons:   hallUp,
		HallDownButtons: hallDown,
	}
}

// Status defines the status of a button.
// All buttons of the same type on the same floor are considered equal,
// and as long as the elevator is online will behave the exact same way.
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

// LiftStatus defines the publicly available information about the elevators in the cluster.
type LiftStatus struct {
	ID                         string
	LastFloor                  uint
	Direction                  string
	DestinationFloor           uint
	DestinationButtonDirection string
	LastUpdate                 time.Time
}

// DeepCopy safely return a copy of the elevator.
func (e *LiftStatus) DeepCopy() LiftStatus {
	return LiftStatus{
		ID:                         e.ID,
		LastFloor:                  e.LastFloor,
		Direction:                  e.Direction,
		DestinationFloor:           e.DestinationFloor,
		DestinationButtonDirection: e.DestinationButtonDirection,
		LastUpdate:                 e.LastUpdate,
	}
}
