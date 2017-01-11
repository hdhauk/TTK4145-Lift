package main

import (
	"fmt"
	"net"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
)

func main() {
	fmt.Println("Start")
	conn, err := net.Dial("tcp", "localhost:15657")
	if err != nil {
		fmt.Println("Failed to connect to simulator")
	}
	defer conn.Close()
	_, floor := driver.ReadFloor(conn)
	fmt.Println("Elevator is on floor: ", floor)
}
