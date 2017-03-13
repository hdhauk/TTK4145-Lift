package driver

// Lift functions
//==============================================================================
func initHW(port string) {
	// Initialize connection to lift
	if ioInit() != nil {
		cfg.Logger.Fatalln("[ERROR] Failed to connect to the elvator. Make sure everything is turned on and try again!")
	}
	clearAllBtns()
	close(liftConnDoneCh)
	cfg.Logger.Println("[INFO] Hardware initialization complete")
}

func setMotorDirHW(dir string) {
	switch dir {
	case stop:
		ioWriteAnalog(motor, 0)
	case up:
		ioClearBit(motorDirDown)
		ioWriteAnalog(motor, 2800)
	case down:
		ioSetBit(motorDirDown)
		ioWriteAnalog(motor, 2800)
	}
}

func setBtnLEDHW(btn Btn, active bool) {
	if active {
		ioSetBit(lampChannelMatrix[btn.Floor][int(btn.Type)])
	} else {
		ioClearBit(lampChannelMatrix[btn.Floor][int(btn.Type)])
	}
}

func setFloorLEDHW(floor int) {
	// Check input validity
	if floor < 0 || floor >= numFloors {
		cfg.Logger.Printf("[Error] Floor %d out of range! No floor indicator will be set.\n", floor)
	}

	// Binary encoding. One light must always be on.
	if floor&0x02 > 0 {
		ioSetBit(floorLED1)
	} else {
		ioClearBit(floorLED1)
	}

	if floor&0x01 > 0 {
		ioSetBit(floorLED2)
	} else {
		ioClearBit(floorLED2)
	}
}

func setDoorLEDHW(isOpen bool) {
	if isOpen {
		ioSetBit(doorOpenLED)
	} else {
		ioClearBit(doorOpenLED)
	}
}

func readOrderBtnHW(btn Btn) bool {
	if ioReadBit(buttonChannelMatrix[btn.Floor][int(btn.Type)]) {
		return true
	}
	return false
}

func readFloorHW() (atFloor bool, floor int) {
	if ioReadBit(sensorFloor1) {
		return true, 0
	} else if ioReadBit(sensorFloor2) {
		return true, 1
	} else if ioReadBit(sensorFloor3) {
		return true, 2
	} else if ioReadBit(sensorFloor4) {
		return true, 3
	} else {
		return false, -1
	}
}

func setStopLamp(active bool) {
	if active {
		ioSetBit(stopLED)
	} else {
		ioClearBit(stopLED)
	}
}

func getObstructionSignal() bool {
	return ioReadBit(obstruct)
}

func getStopSignal() bool {
	return ioReadBit(stopBtn)
}
