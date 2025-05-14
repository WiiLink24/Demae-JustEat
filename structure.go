package main

import (
	"github.com/WiiLink24/nwc24"
	"net/http"
)

type XMLType int

const (
	Normal XMLType = iota
	MultipleRootNodes
	// None is the type used for images
	None
)

// Response describes the inner response format, along with common fields across requests.
type Response struct {
	ResponseFields      any
	hasError            bool
	wiiNumber           nwc24.WiiNumber
	request             *http.Request
	writer              *http.ResponseWriter
	isMultipleRootNodes bool
}
