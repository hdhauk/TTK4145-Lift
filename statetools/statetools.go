package statetools

import "bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"

// GetAssignment calulates the elevator with the lowest cost for a the provided state and order.
func GetAssignment(s globalstate.State, floor int, dir string) string {
	return ""
}

// AddOrder adds an order to the provided state. NB: Only for use on local states.
func AddOrder(s globalstate.State, floor int, dir string) {

}

// RemoveOrder removes an order from the provided state. NB: Only for use on local states.
func RemoveOrder(s globalstate.State, floor int, dir string) {

}
