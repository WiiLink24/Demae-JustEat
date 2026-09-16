package justeat

import (
	"github.com/WiiLink24/DemaeJustEat/demae"
)

var (
	ErrNotLinked           = demae.NewSentryError("Please link your Just Eat Account with\nyour WiiLink Account. Follow the guide\nat https://wiilink.ca/guide/just-eat", false)
	ErrInvalidCountry      = demae.NewSentryError("Your Wii's country does not support\nJust Eat", false)
	ErrNoMenuAvailable     = demae.NewSentryError("This restaurant isn't taking orders\nright now. Please try again later.", false)
	ErrBasketItemNotFound  = demae.NewSentryError("Could not find that item in your\nbasket. Please refresh your basket\nand try again.", false)
	ErrNoBasketProductIds  = demae.NewSentryError("Could not remove that item.\nPlease try again.", true)
	ErrBasketRemovalFailed = demae.NewSentryError("Could not remove that item.\nPlease try again.", true)
	ErrBasketEditFailed    = demae.NewSentryError("Could not update your basket.\nPlease try again.", true)
)
