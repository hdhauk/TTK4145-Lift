package statetools

import (
	"testing"
	"time"
)

func TestCostBasic(t *testing.T) {

	var s = State{
		Nodes: make(map[string]LiftStatus),
	}

	s.Nodes["192.168.0.1:80"] = LiftStatus{
		ID:                         "192.168.0.1:80",
		LastFloor:                  1,
		Direction:                  "stop",
		DestinationFloor:           1,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-1 * time.Second),
	}
	s.Nodes["192.168.0.2:80"] = LiftStatus{
		ID:                         "192.168.0.2:80",
		LastFloor:                  2,
		Direction:                  "up",
		DestinationFloor:           3,
		DestinationButtonDirection: "down",
		LastUpdate:                 time.Now().Add(-2 * time.Second),
	}
	s.Nodes["192.168.0.3:80"] = LiftStatus{
		ID:                         "192.168.0.3:80",
		LastFloor:                  0,
		Direction:                  "stop",
		DestinationFloor:           0,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-7 * time.Second),
	}

	want := "192.168.0.1:80"
	got := CostFunction(s, 2, "up")
	if want != got {
		t.Fatalf("Did not get correct lift: Got = %s, Want = %s", got, want)
	}

}
