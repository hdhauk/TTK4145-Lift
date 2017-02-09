package driver

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Package global channels
var liftConnDone chan bool
var floorDstCh chan int
var btnPressCh chan Btn
var floorDetectCh chan int
var apFloorCh chan int

// Init intializes the driver, and return an error if unable to connect to
// the driver or the simulator.
func Init(c Config, done chan struct{}) error {
	// Set configuration
	if err := setConfig(c); err != nil {
		cfg.Logger.Printf("Failed to set driver configuration: %v", err)
		return err
	}

	// Assign either hardware or simulator functions to driver handle
	if c.SimMode == false {
		driver.init = initHW
		driver.setMotorDir = setMotorDirHW
		driver.setBtnLED = setBtnLEDHW
		driver.setFloorLED = setFloorLEDHW
		driver.setDoorLED = setDoorLEDHW
		driver.readOrderBtn = readOrderBtnHW
		driver.readFloor = readFloorHW
	}

	// Initialize channels
	liftConnDone = make(chan bool)
	btnPressCh = make(chan Btn, 4)
	floorDetectCh = make(chan int)
	apFloorCh = make(chan int)
	floorDstCh = make(chan int)

	// Spawn workers
	go btnScan(btnPressCh)
	go floorDetect(floorDetectCh)
	go btnPressHandler(btnPressCh)
	go floorDetectHandler(floorDetectCh, apFloorCh)
	go autoPilot(apFloorCh, done)
	go driver.init(cfg.SimPort)

	// Block until stack unwind
	select {}
}

// Default config
var cfg = Config{
	SimMode: true,
	SimPort: "53566",
	Floors:  4,
	OnFloorDetect: func(f int) {
		fmt.Printf("onFloorDetect callback not set! Floor: %v\n", f)
	},
	OnNewDirection: func(dir string) {
		fmt.Printf("onNewDirection callback not set! Dir: %v\n", dir)
	},
	OnBtnPress: func(b Btn) {
		fmt.Printf("onBtnPress callback not set! Type: %v, Floor: %v\n", b.Type, b.Floor)
	},
	OnDstReached: func(f int) { fmt.Printf("OnDstReached callback not set! Floor: %v\n", f) },
	Logger:       log.New(os.Stdout, "driver-default-debugger:", log.Lshortfile|log.Ltime),
}

// Config defines the properties of the elevator and callbacks to the following events
//  * OnFloorDetect - The elevator just reached a floor. May or may not stop there.
//  * OnNewDirection - The elevator either stopped or started moving in either direction.
//  * OnBtnPress - A button have been depressed.
type Config struct {
	SimMode        bool
	SimPort        string
	Floors         int
	OnFloorDetect  func(floor int)
	OnNewDirection func(direction string)
	OnDstReached   func(floor int)
	OnBtnPress     func(b Btn)
	Logger         *log.Logger
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

// Config helper functions
func setConfig(c Config) error {
	if c.Logger != nil {
		cfg.Logger = c.Logger
	}

	// Set simulator port
	if c.SimMode {
		if err := validatePort(c.SimPort); err != nil {
			return err
		}
	}
	cfg.SimPort = c.SimPort

	// Set floornumber
	if c.Floors < 0 {
		cfg.Logger.Printf("negative number of floors (%v) not supported\n", c.Floors)
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
		cfg.Logger.Printf("port validation failed. Unable to parse the portnumber: %v\n", port)
		return err
	}
	if i < 1024 || i > 65535 {
		cfg.Logger.Printf("port %d not in valid range range (1024-65553)\n", i)
		return fmt.Errorf("port %d not in valid range range (1024-65553)", i)
	}
	return nil
}
