package justeat

import "errors"

var (
	NotFulfillable    = errors.New("not fulfillable")
	PaypalUnavailable = errors.New("paypal unavailable")
	PaypalURLFailed   = errors.New("creating paypal url failed")
)
