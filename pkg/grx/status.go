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
