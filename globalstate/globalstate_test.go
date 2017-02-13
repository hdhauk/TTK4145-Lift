package globalstate

import "testing"

// func TestInitAndLiftUpdateLocally(t *testing.T) {
// 	cfg := Config{
// 		RaftPort: 8000,
// 		OwnIP:    "localhost",
// 	}
//
// 	if err := Init(cfg); err != nil {
// 		t.Fatalf("Failed to initalize:%s", err.Error())
// 	}
//
// 	lu := LiftStatusUpdate{
// 		Floor: 2,
// 		Dst:   3,
// 		Dir:   "up",
// 	}
//
// 	if err := UpdateLiftStatus(lu); err != nil {
// 		t.Fatalf("Failed to get state: %v", err.Error())
// 	}
//
// 	time.Sleep(4 * time.Second)
//
// 	state, err := GetState()
// 	if err != nil {
// 		t.Fatalf("Failed to get state: %v", err.Error())
// 	}
//
// 	myID := cfg.OwnIP + ":" + strconv.Itoa(cfg.RaftPort)
// 	if state.Nodes[myID].Destination != lu.Dst {
// 		t.Fatalf("Node destination not set correctly! Expected: %v, but got: %v", lu.Dst, state.Nodes[myID].Destination)
// 	} else if state.Nodes[myID].Direction != lu.Dir {
// 		t.Fatalf("Node direction not set correctly! Expected: %v, but got: %v", lu.Dir, state.Nodes[myID].Direction)
// 	} else if state.Nodes[myID].LastFloor != lu.Floor {
// 		t.Fatalf("Node floor not set correctly! Expected: %v, but got: %v", lu.Floor, state.Nodes[myID].LastFloor)
// 	}
// }

// func TestInitAndLiftUpdateLocallyWithManyButtonPresses(t *testing.T) {
// 	cfg := Config{
// 		RaftPort: 8003,
// 		OwnIP:    "localhost",
// 	}
//
// 	if err := Init(cfg); err != nil {
// 		t.Fatalf("Failed to initalize:%s", err.Error())
// 	}
//
// 	lu := LiftStatusUpdate{
// 		CurrentFloor: 2,
// 		CurrentDir:   "up",
// 		DstFloor:   3,
// 		Dst
// 	}
//
// 	if err := UpdateLiftStatus(lu); err != nil {
// 		t.Fatalf("Failed to get state: %v", err.Error())
// 	}
// 	bsu1 := ButtonStatusUpdate{Floor: 0, Dir: "up", Status: BtnStateUnassigned}
// 	bsu2 := ButtonStatusUpdate{Floor: 1, Dir: "down", Status: BtnStateUnassigned}
// 	bsu3 := ButtonStatusUpdate{Floor: 1, Dir: "up", Status: BtnStateUnassigned}
// 	bsu4 := ButtonStatusUpdate{Floor: 2, Dir: "down", Status: BtnStateUnassigned}
// 	bsu5 := ButtonStatusUpdate{Floor: 2, Dir: "up", Status: BtnStateUnassigned}
// 	bsu6 := ButtonStatusUpdate{Floor: 3, Dir: "down", Status: BtnStateUnassigned}
//
// 	go UpdateButtonStatus(bsu1)
// 	go UpdateButtonStatus(bsu2)
// 	go UpdateButtonStatus(bsu3)
// 	go UpdateButtonStatus(bsu4)
// 	go UpdateButtonStatus(bsu5)
// 	go UpdateButtonStatus(bsu6)
//
// 	time.Sleep(4 * time.Second)
//
// 	gotState, err := GetState()
// 	if err != nil {
// 		t.Fatalf("Failed to get state: %v", err.Error())
// 	}
// 	wantState := NewState(4)
// 	n := LiftStatus{
// 		ID:          cfg.OwnIP + ":" + strconv.Itoa(cfg.RaftPort),
// 		LastFloor:   2,
// 		Destination: 3,
// 		Direction:   "up",
// 	}
// 	wantState.Nodes[n.ID] = n
// 	s := Status{LastStatus: BtnStateUnassigned}
// 	wantState.HallUpButtons["0"] = s
// 	wantState.HallUpButtons["1"] = s
// 	wantState.HallUpButtons["2"] = s
// 	wantState.HallUpButtons["3"] = s
// 	wantState.HallDownButtons["0"] = s
// 	wantState.HallDownButtons["1"] = s
// 	wantState.HallDownButtons["2"] = s
// 	wantState.HallDownButtons["3"] = s
// 	compareState(&gotState, wantState, t)
//
// }

func compareState(got, want *State, t *testing.T) {
	for k := range got.Nodes {
		if got.Nodes[k] != want.Nodes[k] {
			t.Fatalf("Expected Nodes[%s] = %+v, but got Nodes[%s] = %+v", k, want.Nodes[k], k, got.Nodes[k])
		}
	}
	for k := range got.HallUpButtons {
		if got.HallUpButtons[k].LastStatus != want.HallUpButtons[k].LastStatus {
			t.Fatalf("Expected HallUpButtons[%s] = %+v, but got HallUpButtons[%s] = %+v", k, want.HallUpButtons[k], k, got.HallUpButtons[k])
		}
	}
	for k := range got.HallDownButtons {
		if got.HallDownButtons[k].LastStatus != want.HallDownButtons[k].LastStatus {
			t.Fatalf("Expected HallDownButtons[%s] = %+v, but got HallDownButtons[%s] = %+v", k, want.HallDownButtons[k], k, got.HallDownButtons[k])
		}
	}

}
