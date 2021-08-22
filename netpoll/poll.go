package netpoll

// Mode represents the read/write mode.
type Mode int

const (
	// READ is the read mode.
	READ Mode = 1 << iota
	// WRITE is the write mode.
	WRITE
)

// Event represents the poll event for the poller.
type Event struct {
	// Fd is a file descriptor.
	Fd int
	// Mode represents the event mode.
	Mode Mode
}
