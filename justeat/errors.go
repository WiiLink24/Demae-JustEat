package justeat

import "errors"

var (
	NotLinked      = errors.New("please link your Just Eat Account with\nyour WiiLink Account")
	InvalidCountry = errors.New("your Wii's country does not support\nJust Eat")

	NoSentryErrors = []error{NotLinked, InvalidCountry}
)
