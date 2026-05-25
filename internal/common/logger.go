package common

import "log"

var debugEnabled bool

func SetDebug(enabled bool) {
	debugEnabled = enabled
}

func Debugf(format string, args ...any) {
	if !debugEnabled {
		return
	}
	log.Printf(format, args...)
}
