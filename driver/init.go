package driver

import (
	"fmt"
	"strconv"
)

// Default config
var cfg = Config{
	SimMode:        true,
	SimPort:        "53566",
	Floors:         4,
	OnFloorDetect:  func(f int) { fmt.Println("onFloorDetect callback not set!") },
	OnNewDirection: func(dir string) { fmt.Println("onNewDirection callback not set!") },
	OnBtnPress:     func(btnType string, floor int) { fmt.Println("onBtnPress callback not set!") },
}

// Config defines the properties of the elevator and callbacks to the following
// events:
//  * OnFloorDetect - The elevator just reached a floor. May or may not stop there.
//  * OnNewDirection - The elevator either stopped or started moving in either direction.
//  * OnBtnPress - A button have been depressed.
type Config struct {
	SimMode        bool
	SimPort        string
	Floors         int
	OnFloorDetect  func(floor int)
	OnNewDirection func(direction string)
	OnBtnPress     func(btnType string, floor int)
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
	if cfg.SimMode == false {
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
	if c.SimMode {
		if err := validatePort(c.SimPort); err != nil {
			return err
		}
	}
	if c.Floors < 0 {
		return fmt.Errorf("negative number of floors (%v) not supported", c.Floors)
	}

	// Check set provided callbacks
	if c.OnFloorDetect != nil {
		cfg.OnFloorDetect = c.OnFloorDetect
	}
	if c.OnBtnPress != nil {
		cfg.OnBtnPress = c.OnBtnPress
	}
	if c.OnNewDirection != nil {
		cfg.OnNewDirection = c.OnNewDirection
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
