package demae

type SentryError struct {
	s string
}

func NewSentryError(s string) error {
	return &SentryError{s: s}
}

func (s *SentryError) Error() string {
	return s.s
}
