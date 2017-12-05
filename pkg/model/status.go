package model

// Status status
type Status int

// Circuit circuit status
type Circuit int

const (
	// Down backend server down status
	Down = Status(0)
	// Up backend server up status
	Up = Status(1)
)

const (
	// CircuitOpen Circuit open status
	CircuitOpen = Circuit(0)
	// CircuitHalf Circuit half status
	CircuitHalf = Circuit(1)
	// CircuitClose Circuit close status
	CircuitClose = Circuit(2)
)
