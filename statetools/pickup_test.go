package statetools

import (
	"testing"
	"time"
)

func TestPickup(t *testing.T) {

	var s = State{
		HallDownButtons: make(map[string]Status),
		HallUpButtons:   make(map[string]Status),
	}

	s.HallUpButtons["1"] = Status{
		AssignedTo: "192.168.0.1:80",
		LastStatus: "assigned",
		LastChange: time.Now().Add(-2 * time.Second),
	}
	s.HallUpButtons["2"] = Status{
		AssignedTo: "192.168.0.1:80",
		LastStatus: "done",
		LastChange: time.Now().Add(-2 * time.Second),
	}

	if !ShouldStopAndPickup(s, 1, "up") {
		t.Errorf("Should have stopped, but did not.")
	}
	if ShouldStopAndPickup(s, 2, "up") {
		t.Errorf("Did stop but shouldn't.")
	}

}
