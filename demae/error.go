package demae

type SentryError struct {
	s      string
	Report bool
}

func NewSentryError(s string, report bool) error {
	return &SentryError{s: s, Report: report}
}

func (s *SentryError) Error() string {
	return s.s
}
