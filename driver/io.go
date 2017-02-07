package driver

/*
	This package is in large part based on the hardware interface by Morten Fyhn (github.com/mortenfyhn)
	--> github.com/mortenfyhn/TTK4145-Heis/blob/master/Lift/src/hw/channels.go
	--> github.com/mortenfyhn/TTK4145-Heis/blob/master/Lift/src/hw/lift.go
	--> github.com/mortenfyhn/TTK4145-Heis/blob/master/Lift/src/hw/io.go

	It is however refactored to comply with the Go guidelines for best practices,
	especially with respect to naming convetions, and some functions are modified
	to be more compatable with the rest of the driver package.
*/

/*
#cgo LDFLAGS: -lcomedi -lm
#include "c_io.h"
*/
import "C"
import "fmt"

func ioInit() error {
	// Initialize hardware
	if int(C.io_init()) == 0 {
		return fmt.Errorf(`unable to initialize hardware driver.
			Make sure everything is turned on and connected`)
	}

	// Turn off all lights
	return nil
}

var lampChannelMatrix = [numFloors][numButtons]int{
	{upLED1, downLED1, cmdLED1},
	{upLED2, downLED2, cmdLED2},
	{upLED3, downLED3, cmdLED3},
	{upLED4, downLED4, cmdLED4},
}
var buttonChannelMatrix = [numFloors][numButtons]int{
	{btnUp1, btnDown1, btnCmd1},
	{btnUp2, btnDown2, btnCmd2},
	{btnUp3, btnDown3, btnCmd3},
	{btnUp4, btnDown4, btnCmd4},
}

const numButtons = 3
const numFloors = 4

// In port 4
const (
	port4    = 3
	obstruct = (0x300 + 23)
	stopBtn  = (0x300 + 22)
	btnCmd1  = (0x300 + 21)
	btnCmd2  = (0x300 + 20)
	btnCmd3  = (0x300 + 19)
	btnCmd4  = (0x300 + 18)
	btnUp1   = (0x300 + 17)
	btnUp2   = (0x300 + 16)
)

// In port 1
const (
	port1        = 2
	btnDown2     = (0x200 + 0)
	btnUp3       = (0x200 + 1)
	btnDown3     = (0x200 + 2)
	btnDown4     = (0x200 + 3)
	sensorFloor1 = (0x200 + 4)
	sensorFloor2 = (0x200 + 5)
	sensorFloor3 = (0x200 + 6)
	sensorFloor4 = (0x200 + 7)
)

// Out port 3
const (
	port3        = 3
	motorDirDown = (0x300 + 15)
	stopLED      = (0x300 + 14)
	cmdLED1      = (0x300 + 13)
	cmdLED2      = (0x300 + 12)
	cmdLED3      = (0x300 + 11)
	cmdLED4      = (0x300 + 10)
	upLED1       = (0x300 + 9)
	upLED2       = (0x300 + 8)
)

// Out port 2
const (
	port2       = 3
	downLED2    = (0x300 + 7)
	upLED3      = (0x300 + 6)
	downLED3    = (0x300 + 5)
	downLED4    = (0x300 + 4)
	doorOpenLED = (0x300 + 3)
	floorLED2   = (0x300 + 1)
	floorLED1   = (0x300 + 0)
)

// Out port 0
const (
	port0 = 1
	motor = (0x100 + 0)
)

// Non-existing ports = (for alignment)
const (
	btnDown1 = -1
	btnUp4   = -1
	downLED1 = -1
	upLED4   = -1
)

func ioSetBit(channel int) {
	C.io_set_bit(C.int(channel))
}

func ioClearBit(channel int) {
	C.io_clear_bit(C.int(channel))
}

func ioWriteAnalog(channel int, value int) {
	C.io_write_analog(C.int(channel), C.int(value))
}

func ioReadBit(channel int) bool {
	return int(C.io_read_bit(C.int(channel))) != 0
}

func ioReadAnalog(channel int) int {
	return int(C.io_read_analog(C.int(channel)))
}
