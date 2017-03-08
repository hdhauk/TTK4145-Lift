/*
Package driver provides control of both simulated and actual lifts.
The package also provide functionality for handling internal orders, as well as taking external orders.
*/
package driver

import "fmt"

// GoToFloor sends the lift carriage to the desired floor and stop there,
// unless it is stopped before arriving at its destination.
// A second call to the function will void the previous order if the carriage
// haven't reached its destination.
func GoToFloor(floor int, dir string) {
	if floor > cfg.Floors-1 || floor < 0 {
		cfg.Logger.Printf("%s[ERROR] Invalid floor requested: %v%s\n", yellow, floor, white)
		return
	}
	if floor >= 0 {
		floorDstCh <- dst{floor: floor, dir: dir}
	}
}

// StopForPickup can be called if the lift should stop in the next floor,
// to pick someone up.
func StopForPickup(f int, d string) {
	stopForPickupCh <- dst{f, d}
}

// BtnLEDClear turns off the LED in the provided button.
func BtnLEDClear(b Btn) {
	if err := validateButton(b); err != nil {
		cfg.Logger.Printf("%s[ERROR] Invalid button: %s%s", yellow, err.Error(), white)
		return
	}
	// TODO: Check for race conditions
	driverHandle.setBtnLED(b, false)
}

// BtnLEDSet turns on the LED in the provided button.
func BtnLEDSet(b Btn) {
	if err := validateButton(b); err != nil {
		cfg.Logger.Printf("%s[ERROR] Invalid button: %s%s", yellow, err.Error(), white)
		return
	}
	driverHandle.setBtnLED(b, true)
}

// BtnType defines the 3 types of buttons that are in use. In order to use the
// correct integer, the types are available as constants.
type BtnType int

// Button type constants
const (
	// HallUp is located outside of the lift.
	HallUp BtnType = iota
	// HallDown is located outside of the lift.
	HallDown
	// Cab is located inside of the lift.
	Cab
)

// Motor direction constants
const (
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
		return "up"
	case HallDown:
		return "down"
	case Cab:
		return "Cab"
	}
	return ""
}

func validateButton(b Btn) error {
	if b.Floor > cfg.Floors-1 || b.Floor < 0 {
		return fmt.Errorf("floor not in range [ %d - %d ]", 0, cfg.Floors-1)
	}
	if b.Floor == 0 && b.Type == HallDown {
		return fmt.Errorf("no down button at ground floor")
	}
	if b.Floor == cfg.Floors-1 && b.Type == HallUp {
		return fmt.Errorf("no up button at top floor")
	}
	return nil
}

// Stop immediately stop the elevator wherever it is and turns all LEDs off.
func Stop() {
	driverHandle.setMotorDir(stop)
	clearAllBtns()
}
