package notification

import "fmt"

type ServerStatus uint8

const (
	Undefined ServerStatus = iota
	Online
	Offline
	ShuttingDown
	Saturated
)

func (s ServerStatus) String() string {
	switch s {
	case Undefined:
		return "Undefined"
	case Online:
		return "Online"
	case Offline:
		return "Offline"
	case ShuttingDown:
		return "Shutting Down"
	case Saturated:
		return "Saturated"
	}
	return "Unknown"
}

type ServerState struct {
	Code    ServerStatus
	Message string
}

func (s *ServerState) String() string {
	return fmt.Sprintf("Status: %s\nMessage: %s", s.Code, s.Message)
}
