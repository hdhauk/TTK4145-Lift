package statetools

import (
	"testing"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
)

func TestCostBasic(t *testing.T) {

	var s = globalstate.State{
		Nodes: make(map[string]globalstate.LiftStatus),
	}

	s.Nodes["192.168.0.1:80"] = globalstate.LiftStatus{
		ID:                         "192.168.0.1:80",
		LastFloor:                  1,
		Direction:                  "STOP",
		DestinationFloor:           1,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-1 * time.Second),
	}
	s.Nodes["192.168.0.2:80"] = globalstate.LiftStatus{
		ID:                         "192.168.0.2:80",
		LastFloor:                  2,
		Direction:                  "STOP",
		DestinationFloor:           2,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-2 * time.Second),
	}
	s.Nodes["192.168.0.3:80"] = globalstate.LiftStatus{
		ID:                         "192.168.0.3:80",
		LastFloor:                  3,
		Direction:                  "STOP",
		DestinationFloor:           3,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-1 * time.Second),
	}
	s.Nodes["192.168.0.4:80"] = globalstate.LiftStatus{
		ID:                         "192.168.0.4:80",
		LastFloor:                  0,
		Direction:                  "STOP",
		DestinationFloor:           0,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-1 * time.Second),
	}

	want := "192.168.0.1:80"
	got := CostFunction(s, 1, "up")
	if want != got {
		t.Fatalf("Did not get correct lift: Got = %s, Want = %s", got, want)
	}

}
