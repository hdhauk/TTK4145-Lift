package driver

import (
	"fmt"

	logging "github.com/op/go-logging"
)

var simMode = false
var initalized = false
var logger *logging.Logger

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

// Init intializes an elevator or simulated elevator.
func Init(simPort string, loggerCallBack *logging.Logger) {
	logger = loggerCallBack
	if simPort != "" {
		simMode = true
		initalized = true
		initSim(simPort)
		return
	}
	// C.elev_init(C.ET_Comedi)
}

func isInitialized() bool {
	return initalized
}

// SetMotorDir either stops the motor or sets it in motion in either direction.
func SetMotorDir(dir string) {
	if !isInitialized() {
		fmt.Println("Driver not initialized!")
	}
	switch simMode {
	case true:
		setMotorDirSim(dir)
	case false:
		// setMotorDir(dir)
	}
}

// SetBtnLED sets the indicator lights on buttons.
func SetBtnLED(btn Btn, active bool) {
	if !isInitialized() {
		fmt.Println("Driver not initialized!")
	}
	switch simMode {
	case true:
		setBtnLEDSim(btn, active)
	case false:
		// setBtnLED(btn, active)
	}
}

// SetFloorLED sets the floor indicator LED
func SetFloorLED(floor int) {
	if !isInitialized() {
		fmt.Println("Driver not initialized!")
	}
	switch simMode {
	case true:
		setFloorLEDSim(floor)
	case false:
		// setFloorLED(floor)
	}
}

// SetDoorLED sets the door-open LED
func SetDoorLED(isOpen bool) {
	if !isInitialized() {
		fmt.Println("Driver not initialized!")
	}
	switch simMode {
	case true:
		setDoorLEDSim(isOpen)
	case false:
		// setDoorLED(isOpen)
	}
}

// SetStopLED sets the stop-light LED
func SetStopLED(active bool) {
	//TODO
}

// ReadOrderBtn return true if the button btn is *currently* being pressed, and
// otherwise false.
func ReadOrderBtn(btn Btn) bool {
	if !isInitialized() {
		fmt.Println("Driver not initialized!")
	}
	switch simMode {
	case true:
		return readOrderBtnSim(btn)
	case false:
		// return  readOrderBtn(btn)
	}
	return false
}

// ReadFloor returns the position or status of the elevator.
func ReadFloor() (atFloor bool, floor int) {
	if !isInitialized() {
		fmt.Println("Driver not initialized!")
	}
	switch simMode {
	case true:
		return readFloorSim()
	case false:
		// return  readFloor()
	}
	return false, -1
}
