package statetools

// State is identical with the State in the globalstate
type State struct{}

// GetAssignment calulates the elevator with the lowest cost for a the provided state and order.
func GetAssignment(s State, floor int, dir string) string {
	return ""
}

// AddOrder adds an order to the provided state. NB: Only for use on local states.
func AddOrder(s State, floor int, dir string) {

}

// RemoveOrder removes an order from the provided state. NB: Only for use on local states.
func RemoveOrder(s State, floor int, dir string) {

}
