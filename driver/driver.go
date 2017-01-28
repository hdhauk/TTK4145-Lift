/*
Package driver provides control of both simulated and actual elevators.
The package also usure that the floor-indicator always show the correct floor,
and that the carrige always have a closed door unless stationary at a floor.
*/
package driver

var floorDstCh = make(chan int)

// GoToFloor sends the elevator carrige to the desired floor and stop there,
// unless it is stopped before arriving at its destination.
// A second call to the function will void the previous order if the carrige
// haven't reached its destination.
func GoToFloor(floor int) {
	if floor >= 0 {
		floorDstCh <- floor
	}
}

// StopAtNextFloor safely stop the elevator the next time it is in a floor and
// open the door.
func StopAtNextFloor() {

}

// BtnType ..
type BtnType int

const (
	// HallUp is located outside of the elevator
	HallUp BtnType = iota
	// HallDown is located outside of the elevator
	HallDown
	// Cab is located inside of the elevator
	Cab
)

// Btn defines a custom type representing a hardware-button. Floor may be nil.
type Btn struct {
	Floor int
	Type  BtnType
}

func (bt *BtnType) String() string {
	switch *bt {
	case HallUp:
		return "HallUp"
	case HallDown:
		return "HallDown"
	case Cab:
		return "Cab"
	}
	return ""
}
