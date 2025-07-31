package justeat

import "errors"

var (
	NotLinked         = errors.New("please link your Just Eat Account with\nyour WiiLink Account")
	NotFulfillable    = errors.New("not fulfillable")
	PaypalUnavailable = errors.New("paypal unavailable")
	PaypalURLFailed   = errors.New("creating paypal url failed")
)
