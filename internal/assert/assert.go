// Package assert provides a simple assertion mechanism for Go programs.
package assert

import "log"

func Assert(condition bool, message string) {
	if !condition {
		log.Fatalln(message)
	}
}
