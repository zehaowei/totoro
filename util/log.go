package util

import "log"

// Debugging
const logTier = 2

func PrintInfo(format string, a ...interface{}) (n int, err error) {
	if logTier > 1 {
		log.Printf(format, a...)
	}
	return
}

func PrintErr(format string, a ...interface{}) (n int, err error) {
	if logTier > 0 {
		log.Printf(format, a...)
	}
	return
}