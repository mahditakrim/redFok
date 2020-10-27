package main

import (
	"fmt"
)

// errScope is the struct that we use to package errors and their scopes together
// scope is usually the name of the func that has caused the err
// err is the error that has been occurred
type errScope struct {
	scope string
	err   error
}

//We may want logging processes here . . .

// logError just logs errors yet for the development progress.
// it gets a scope which is the scope of code that the err is came from.
// it gets an error which is the error that just happened.
func logError(scope string, err error) {

	go playBeep()
	fmt.Println("Error: ", scope, "---", err)
}
