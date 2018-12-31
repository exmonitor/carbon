package service

type Service struct {
	ID            int    // from Service table
	Type          int    // from Service table
	FailThreshold int    // from Service table
	Interval      int    // from Service table
	Host          string // from Service table
	Target        string // from Host table via FK join on Host

	Metadata string // from Service table
}

func (s *Service) ServiceTypeString() string {
	switch s.Type {
	case 1:
		return "http"
	case 2:
		return "tcp"
	case 3:
		return "icmp"
	default:
		return "unknown"
	}
}
