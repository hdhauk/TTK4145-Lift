package driver

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type dst struct {
	floor int
	dir   string
}

// Package global channels
var liftConnDoneCh chan bool
var floorDstCh chan dst
var stopForPickupCh chan dst
var btnPressCh chan Btn
var floorDetectCh chan int
var apFloorCh chan int

// Init intializes the driver, and return an error on the done-channel
// if unable to initialize the driver
func Init(c Config, done chan error) {
	// Set configuration
	if err := setConfig(c); err != nil {
		cfg.Logger.Printf("Failed to set driver configuration: %v", err)
		done <- err
		return
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
	liftConnDoneCh = make(chan bool)
	btnPressCh = make(chan Btn, c.Floors)
	floorDetectCh = make(chan int, c.Floors)
	stopForPickupCh = make(chan dst)
	apFloorCh = make(chan int)
	floorDstCh = make(chan dst, c.Floors)

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

// Default config (may be partially or completely overwritten)
var cfg = Config{
	SimMode:     true,
	SimPort:     "53566",
	Floors:      4,
	OnNewStatus: func(f int, dir string, d int, dd string) { fmt.Println("OnNewStatus callback not set!") },
	OnBtnPress: func(b Btn) {
		fmt.Printf("onBtnPress callback not set! Type: %v, Floor: %v\n", b.Type, b.Floor)
	},
	OnDstReached: func(b Btn) { fmt.Printf("OnDstReached callback not set! Floor: %v\n", b.Floor) },
	Logger:       log.New(os.Stdout, "driver-default-debugger:", log.Lshortfile|log.Ltime),
}

// Config defines the configuration for the driver.
type Config struct {
	SimMode      bool
	SimPort      string
	Floors       int
	OnNewStatus  func(floor int, dir string, dstFloor int, dstDir string)
	OnDstReached func(b Btn)
	OnBtnPress   func(b Btn)
	Logger       *log.Logger
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

// Update the default config with supplied values
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
	cfg.Floors = c.Floors

	// Check set provided callbacks
	if c.OnNewStatus != nil {
		cfg.OnNewStatus = c.OnNewStatus
	}
	if c.OnBtnPress != nil {
		cfg.OnBtnPress = c.OnBtnPress
	}
	if c.OnDstReached != nil {
		cfg.OnDstReached = c.OnDstReached
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
