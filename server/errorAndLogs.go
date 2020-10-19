package main

import (
	"fmt"
)

//We may want logging processes here . . .

// logError just logs errors yet for the development progress.
// it gets a scope which is the scope of code that the err is came from.
// it gets an error which is the error that just happened.
func logError(scope string, err error) {

	go playBeep()
	fmt.Println("Error: ", scope, "***", err)
}
