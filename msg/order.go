package msg

import "time"

// OrderStatus is a message emitted when an elevator either begin assign
// and order to its internal queue, begin expediting an order, or have
// completed an order.
type OrderStatus struct {
	OrderID string
	Status  string // "RECIEVED", "ASSIGNED", "BEGIN", "COMPLETE"
}

// ExtOrder is an order that get transmitted to the rest of the network.
type ExtOrder struct {
	OrderID string // UUID
	SrcID   string
	Dir     string
	Floor   int
}

// IntOrder is the internal
type IntOrder struct {
	ExtOrder
	OrderStatus
	Created time.Time
}
