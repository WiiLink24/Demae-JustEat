package justeat

import (
	"github.com/WiiLink24/DemaeJustEat/demae"
)

var (
	ErrNotLinked      = demae.NewSentryError("Please link your Just Eat Account with\nyour WiiLink Account")
	ErrInvalidCountry = demae.NewSentryError("Your Wii's country does not support\nJust Eat")
)
