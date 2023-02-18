package grx

type serverStatus uint8

const (
	online serverStatus = iota
	offline
	shuttingDown
)

func (s serverStatus) String() string {
	switch s {
	case online:
		return "online"
	case offline:
		return "offline"
	case shuttingDown:
		return "shutting down"
	}
	return "unknown"
}

type grxStatus uint8

const (
	running grxStatus = iota
	stopped
	stopping
	awaitting
)

func (s grxStatus) String() string {
	switch s {
	case running:
		return "running"
	case stopped:
		return "stopped"
	case stopping:
		return "stopping"
	}
	return "unknown"
}
