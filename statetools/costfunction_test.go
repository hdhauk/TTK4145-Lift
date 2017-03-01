package statetools

import (
	"testing"
	"time"

	. "github.com/hdhauk/TTK4145-Lift/globalstate"
)

func Test_CostBasic(t *testing.T) {

	var s = State{
		Nodes: make(map[string]LiftStatus),
	}

	s.Nodes["192.168.0.1:80"] = LiftStatus{
		ID:                         "192.168.0.1:80",
		LastFloor:                  1,
		Direction:                  "STOP",
		DestinationFloor:           1,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-1 * time.Second),
	}
	s.Nodes["192.168.0.2:80"] = LiftStatus{
		ID:                         "192.168.0.2:80",
		LastFloor:                  2,
		Direction:                  "STOP",
		DestinationFloor:           2,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-2 * time.Second),
	}
	s.Nodes["192.168.0.3:80"] = LiftStatus{
		ID:                         "192.168.0.3:80",
		LastFloor:                  3,
		Direction:                  "STOP",
		DestinationFloor:           3,
		DestinationButtonDirection: "",
		LastUpdate:                 time.Now().Add(-1 * time.Second),
	}
	s.Nodes["192.168.0.4:80"] = LiftStatus{
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

func Test_NoAvailableLifts(t *testing.T) {
	testState := State{
		Nodes: map[string]LiftStatus{
			"192.168.0.0.4:8000": LiftStatus{
				ID:                         "192.168.0.0.4:8000",
				LastFloor:                  3,
				Direction:                  "UP",
				DestinationFloor:           4,
				DestinationButtonDirection: "UP",
				LastUpdate:                 time.Now().Add(-1 * time.Second),
			},
			"192.168.0.0.4:8002": LiftStatus{
				ID:                         "192.168.0.0.4:8004",
				LastFloor:                  2,
				Direction:                  "DOWN",
				DestinationFloor:           0,
				DestinationButtonDirection: "UP",
				LastUpdate:                 time.Now().Add(-1 * time.Second),
			},
			"192.168.0.0.4:8004": LiftStatus{
				ID:                         "192.168.0.0.4:8004",
				LastFloor:                  3,
				Direction:                  "STOP",
				DestinationFloor:           3,
				DestinationButtonDirection: "",
				LastUpdate:                 time.Now().Add(-1 * time.Second),
			},
		},
		HallUpButtons: map[string]Status{
			"0": Status{
				AssignedTo: "192.168.0.0.4:8002",
				LastStatus: "assigned",
				LastChange: time.Now().Add(-3 * time.Second),
			},
			"1": Status{
				AssignedTo: "",
				LastStatus: "unassigned",
				LastChange: time.Now().Add(-3 * time.Second),
			},
			"3": Status{
				AssignedTo: "192.168.0.0.4:8004",
				LastStatus: "assigned",
				LastChange: time.Now().Add(-3 * time.Second),
			},
		},
	}

	want := ""
	got := CostFunction(testState, 3, "up")
	if want != got {
		t.Fatalf("Did not get correct lift: Got = %s, Want = %s", got, want)
	}

}
