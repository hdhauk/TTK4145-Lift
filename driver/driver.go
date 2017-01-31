/*
Package driver provides control of both simulated and actual elevators.
The package also usure that the floor-indicator always show the correct floor,
and that the carrige always have a closed door unless stationary at a floor.
*/
package driver

import "fmt"

// GoToFloor sends the elevator carrige to the desired floor and stop there,
// unless it is stopped before arriving at its destination.
// A second call to the function will void the previous order if the carrige
// haven't reached its destination.
func GoToFloor(floor int) {
	if floor > cfg.Floors-1 || floor < 0 {
		fmt.Println("How about going to a floor that actually exists, hu! ...smartass...")
		return
	}
	if floor >= 0 {
		floorDstCh <- floor
	}
}

// StopAtNextFloor safely stop the elevator the next time it is in a floor and
// open the door.
// func StopAtNextFloor() {
//
// }

// BtnLEDClear turns off the LED in the provided button.
func BtnLEDClear(b Btn) {
	// TODO: Check for race conditions
	driver.setBtnLED(b, false)
}

// BtnLEDSet turns on the LED in the provided button.
func BtnLEDSet(b Btn) {
	// TODO: Check for race conditions
	driver.setBtnLED(b, true)
}

// BtnType defines the 3 types of buttons that are in use. In order to use the
// correct integer, the types are available as constants.
type BtnType int

const (
	// HallUp is located outside of the elevator
	HallUp BtnType = iota
	// HallDown is located outside of the elevator
	HallDown
	// Cab is located inside of the elevator
	Cab

	stop = "STOP"
	up   = "UP"
	down = "DOWN"
)

// Btn defines a custom type representing a hardware-button. Floor may be nil.
type Btn struct {
	Floor int
	Type  BtnType
}

// String exports the integer BtnType to a string (either "HallUP","HallDown" or"Cab")
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
