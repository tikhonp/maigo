package assert

import "log"

func Assert(condition bool, message string) {
	if !condition {
		log.Fatalln(message)
	}
}

