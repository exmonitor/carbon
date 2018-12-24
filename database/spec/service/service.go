package service

type Service struct {
	FailThreshold int
	Host          string
	ID            int
	Interval      int
	Port          int
	ServiceType   int
	Target        string
}

func (s *Service) ServiceTypeString() string {
	switch s.ServiceType {
	case 0:
		return "http/https"
	case 1:
		return "tcp"
	case 2:
		return "icmp"
	default:
		return "unknown"
	}
}
