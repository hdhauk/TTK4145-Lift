package driver

import (
	"fmt"
	"strconv"
)

// Default config
var cfg = Config{
	simMode:        true,
	simPort:        "53566",
	floors:         4,
	onFloorDetect:  func(f int) { fmt.Println("onFloorDetect callback not set!") },
	onNewDirection: func(dir string) { fmt.Println("onNewDirection callback not set!") },
	onBtnPress:     func(btnType string, floor int) { fmt.Println("onBtnPress callback not set!") },
}

// Config defines the properties of the elevator and callbacks to the following
// events:
//  * onFloorDetect - The elevator just reached a floor. May or may not stop there.
//  * onNewDirection - The elevator either stopped or started moving in either direction.
//  * onBtnPress - A button have been depressed.
type Config struct {
	simMode        bool
	simPort        string
	floors         int
	onFloorDetect  func(floor int)
	onNewDirection func(direction string)
	onBtnPress     func(btnType string, floor int)
}

var driver = struct {
	init         func(port string)
	setMotorDir  func(dir string)
	setBtnLED    func(btn Btn, active bool)
	setFloorLED  func(floor int)
	setDoorLED   func(isOpen bool)
	readOrderBtn func(btn Btn) bool
	readFloor    func() (atFloor bool, floor int)
}{
	init:         initSim,
	setMotorDir:  setMotorDirSim,
	setBtnLED:    setBtnLEDSim,
	setFloorLED:  setFloorLEDSim,
	setDoorLED:   setDoorLEDSim,
	readOrderBtn: readOrderBtnSim,
	readFloor:    readFloorSim,
}

// Init intializes the driver, and return an error if unable to connect to
// the driver or the simulator.
func Init(c Config) error {
	// Set configuration
	if err := setConfig(c); err != nil {
		fmt.Printf("Failed to set driver configuration: %v", err)
		return err
	}

	// Assign either hardware or simulator functions to driver handle
	if cfg.simMode == false {
		driver.init = initHW
		driver.setMotorDir = setMotorDirHW
		driver.setBtnLED = setBtnLEDHW
		driver.setFloorLED = setFloorLEDHW
		driver.setDoorLED = setDoorLEDHW
		driver.readOrderBtn = readOrderBtnHW
		driver.readFloor = readFloorHW
	}

	// Spawn workers
	// TODO: What workers do we need here ¯\_(ツ)_/¯

	return nil
}

// Config helper functions
func setConfig(c Config) error {
	if c.simMode {
		if err := validatePort(c.simPort); err != nil {
			return err
		}
	}
	if c.floors < 0 {
		return fmt.Errorf("negative number of floors (%v) not supported", c.floors)
	}

	// Check set provided callbacks
	if c.onFloorDetect != nil {
		cfg.onFloorDetect = c.onFloorDetect
	}
	if c.onBtnPress != nil {
		cfg.onBtnPress = c.onBtnPress
	}
	if c.onNewDirection != nil {
		cfg.onNewDirection = c.onNewDirection
	}
	return nil
}

func validatePort(port string) error {
	i, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	if i < 1024 || i > 65535 {
		return fmt.Errorf("port %d not in valid range range (1024-65553)", i)
	}
	return nil
}
