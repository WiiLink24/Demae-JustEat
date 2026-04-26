package justeat

import (
	"github.com/WiiLink24/DemaeJustEat/demae"
)

var (
	ErrNotLinked      = demae.NewSentryError("Please link your Just Eat Account with\nyour WiiLink Account. Follow the guide\nat https://wiilink.ca/guide/just-eat", false)
	ErrInvalidCountry = demae.NewSentryError("Your Wii's country does not support\nJust Eat", false)
)
