package hw

import (
	"fmt"
	"net"
)

var simPort = "53566"
var txWithResp = make(chan string)
var txWithoutResp = make(chan string)
var rx = make(chan []byte)
var abort = make(chan bool)

func initSim() {
	connStr := fmt.Sprintf("localhost:%s", simPort)
	conn, err := net.Dial("tcp", connStr)
	if err != nil {
		fmt.Println("Failed to connect to simulator:(((")
	}
	defer conn.Close()

	for {
		select {
		case cmd := <-txWithResp:
			fmt.Fprintf(conn, cmd)
			resp := make([]byte, 4)
			fmt.Println("About to listen...")
			conn.Read(resp)
			fmt.Println("Done:)")
			rx <- resp
		case cmd := <-txWithoutResp:
			fmt.Fprintf(conn, cmd)
		}
	}
}

// Helper functions
//==============================================================================
func poll(cmd string) []byte {
	txWithResp <- cmd
	return <-rx
}

func sendCmd(cmd string) {
	txWithoutResp <- cmd
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Hex command generators
//==============================================================================
func cmdMotorDir(dir string) string {
	switch dir {
	case "UP":
		return "\x01\x01\x00\x00"
	case "STOP":
		return "\x01\x00\x00\x00"
	case "DOWN":
		return "\x01\xFF\x00\x00"
	default:
		return ""
	}
}

func cmdBtnLED(btn Btn, active bool) string {
	code := string([]byte{2, byte(btn.Type), byte(btn.Floor), byte(btoi(active))})
	return code
}

func cmdFloorLED(floor int) string {
	return string([]byte{3, byte(floor), 0, 0})
}

func cmdDoorLED(isOpen bool) string {
	return string([]byte{4, byte(btoi(isOpen)), 0, 0})
}

func cmdReadOrderBtn(btn Btn) string {
	return string([]byte{6, byte(btn.Type), byte(btn.Floor), 0})
}

// Emulated elevator functions
//==============================================================================
func setMotorDirSim(dir string) {
	sendCmd("GET " + cmdMotorDir(dir))
}

func setBtnLEDSim(btn Btn, active bool) {
	sendCmd("GET " + cmdBtnLED(btn, active))
}

func setFloorLEDSim(floor int) {
	sendCmd("GET " + cmdFloorLED(floor))
}

func setDoorLEDSim(isOpen bool) {
	sendCmd("GET " + cmdDoorLED(isOpen))
}

func readOrderBtnSim(btn Btn) bool {
	resp := poll("GET " + cmdReadOrderBtn(btn))
	return resp[1] == 1
}

func readFloorSim() (atFloor bool, floor int) {
	resp := poll("GET \x07\x00\x00\x00")
	return (resp[1] != 0), int(resp[2])
}
