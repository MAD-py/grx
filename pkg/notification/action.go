package notification

type ServerAction uint8

const (
	Shutdown ServerAction = iota
	Report
)

func (a ServerAction) String() string {
	switch a {
	case Shutdown:
		return "Shutdown"
	case Report:
		return "Report"
	}
	return "Unknown Server Action"
}
