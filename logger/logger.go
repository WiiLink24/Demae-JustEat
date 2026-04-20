package logger

import (
	"fmt"
	"log"

	"github.com/logrusorgru/aurora/v4"
)

var isDebug bool

func SetDebug(d bool) {
	isDebug = d
}

func Debug(module string, arguments ...any) {
	if !isDebug {
		return
	}

	var finalStr string
	for _, argument := range arguments {
		finalStr += fmt.Sprint(argument)
		finalStr += " "
	}

	log.Printf(aurora.BrightCyan("DEBUG [%s]").String()+": %s", module, finalStr)
}

func Error(module string, arguments ...any) {
	var finalStr string
	for _, argument := range arguments {
		finalStr += fmt.Sprint(argument)
		finalStr += " "
	}

	log.Printf(aurora.BrightRed("ERROR [%s]").String()+": %s", module, finalStr)
}
