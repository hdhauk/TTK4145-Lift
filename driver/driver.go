package driver

import (
	"fmt"
	"net"
)

func sendCommand(c net.Conn, command string) (resp []byte) {
	fmt.Fprintf(c, command)
	tmp := make([]byte, 4)
	c.Read(tmp)
	return tmp
}

func ReadFloor(c net.Conn) (atFloor bool, floor int) {
	answ := sendCommand(c, "GET \x07\x00\x00\x00")
	return (answ[1] != 0), int(answ[2])
}
