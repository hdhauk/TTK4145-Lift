package driver

import (
	"fmt"
	"net"
)

// BtnType ...
type BtnType int

const (
	callUp BtnType = iota
	callDown
	command
)

// Btn defines a custom type representing a hardware-button. Floor may be nil.
type Btn struct {
	Floor int
	Type  BtnType
}

func sendCommand(c net.Conn, command string) (resp []byte) {
	fmt.Fprintf(c, command)
	tmp := make([]byte, 4)
	c.Read(tmp)
	return tmp
}

// Init intializes the elevator and returns any errors
func Init() error {
	// TODO
}

// SetMotorDir either stops the motor or sets it in motion in either direction.
func SetMotorDir(dir string) {
	switch dir {
	case "UP":
		// TODO
		break
	case "DOWN":
		// TODO
		break
	case "STOP":
		// TODO
		break
	}
}

// SetBtnLED sets the indicator lights on buttons.
func SetBtnLED(btn Btn, active bool) {
	// TODO
}

// SetFloorLED sets the floor indicator LED
func SetFloorLED(int floor, active bool) {
	// TODO
}

// SetDoorLED sets the door-open LED
func SetDoorLED(isOpen bool) {
	// TODO
}

// SetStopLED sets the stop-light LED
func SetStopLED(active bool) {
	//TODO
}

// ReadFloor returns the position or status of the elevator.
func ReadFloor(c net.Conn) (atFloor bool, floor int) {
	answ := sendCommand(c, "GET \x07\x00\x00\x00")
	return (answ[1] != 0), int(answ[2])
}
