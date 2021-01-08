package util

import "log"

// Debugging
const logTier = 2

func PrintInfo(format string, a ...interface{}) {
	if logTier > 1 {
		log.Printf(format, a...)
	}
	return
}

func PrintErr(format string, a ...interface{}) {
	if logTier > 0 {
		log.Printf(format, a...)
	}
	return
}