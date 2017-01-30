package driver

func initHW(port string) {

}

func setMotorDirHW(dir string) {

}

func setBtnLEDHW(btn Btn, active bool) {

}

func setFloorLEDHW(floor int) {

}

func setDoorLEDHW(isOpen bool) {

}

func readOrderBtnHW(btn Btn) bool {
	return false
}

func readFloorHW() (atFloor bool, floor int) {
	return false, 0
}
