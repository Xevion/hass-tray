package internal

type Status int

const (
	StatusUnknown Status = iota
	StatusRunning
	StatusStopped
	StatusError
)

type Service interface {
	Status() Status
	State() string
	Connected() bool
	Reload() error
	Pause() error
	Resume() error
}
