package hw

/*
// #include "c-files/io.h"
// #cgo CFLAGS: -std=c99
// #cgo LDFLAGS: -L. -lcomedi -lm
*/
import "C"
import (
	"fmt"
	"time"
)

// NumFloors is the number of floors on test-harware
const NumFloors = 4

// HardwareError is an custom error type for errors conserning the test-harware.
type HardwareError struct {
	What string
	When time.Time
}

func (e HardwareError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func ioInit() error {
	// status := C.io_init()
	// if status == 0 {
	// 	err := HardwareError(
	// 		"Unable to initialize hardware",
	// 		time.Now(),
	// 	)
	// 	fmt.Println(err.Error())
	// 	return err
	// }
	return nil
}
