package justeat

import "errors"

var (
	ErrNotLinked      = errors.New("please link your Just Eat Account with\nyour WiiLink Account")
	ErrInvalidCountry = errors.New("your Wii's country does not support\nJust Eat")
)
