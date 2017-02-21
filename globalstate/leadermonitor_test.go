package globalstate

import "testing"

func Test_GetUnnasignedOrders(t *testing.T) {
	// Create some test-states
	testState1 := NewState(4) // No unnassigned
	testState2 := NewState(4) // Some unnasigned
	testState3 := NewState(4) // Some bad data

	// Create some different buttons statuses
	statusAssigned := Status{LastStatus: BtnStateAssigned}
	statusDone := Status{LastStatus: BtnStateDone}
	statusUnassigned := Status{LastStatus: BtnStateUnassigned}

	// Populate test states
	testState1.HallUpButtons["1"] = statusDone
	testState1.HallUpButtons["0"] = statusAssigned
	testState1.HallDownButtons["3"] = statusDone

	testState2.HallUpButtons["0"] = statusAssigned
	testState2.HallUpButtons["1"] = statusUnassigned
	testState2.HallUpButtons["2"] = statusDone
	testState2.HallUpButtons["3"] = statusUnassigned
	testState2.HallDownButtons["0"] = statusUnassigned
	testState2.HallDownButtons["1"] = statusDone
	testState2.HallDownButtons["2"] = statusDone
	testState2.HallDownButtons["3"] = statusUnassigned

	testState3.HallUpButtons["0"] = Status{LastStatus: "asdas"}
	testState3.HallUpButtons["1"] = statusUnassigned
	testState3.HallUpButtons["2"] = statusDone
	testState3.HallUpButtons["3"] = Status{LastStatus: "statusUnassigned"}
	testState3.HallDownButtons["0"] = statusUnassigned
	testState3.HallDownButtons["1"] = statusDone
	testState3.HallDownButtons["2"] = Status{LastStatus: "___"}
	testState3.HallDownButtons["3"] = Status{LastStatus: ""}

	// Populate test array
	var tests = []struct {
		input *State
		want  []btn
	}{
		{input: testState1, want: []btn{}},
		{input: testState2, want: []btn{btn{1, "up"}, btn{3, "up"}, btn{0, "down"}, btn{3, "down"}}},
		{input: testState3, want: []btn{btn{1, "up"}, btn{0, "down"}}},
	}

	// Run tests
	for _, test := range tests {
		got := getUnassignedOrders(*test.input)
	middleLoop:
		for _, wantOrder := range test.want {
			for _, gotOrder := range got {
				if wantOrder == gotOrder {
					break middleLoop
				}
			}
			t.Errorf("Could not find %+v", wantOrder)
		}
	}
}
