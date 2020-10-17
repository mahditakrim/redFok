package main

import (
	"fmt"
	"log"
)

//We may want logging processes here . . .

// logError just logs errors yet for the development progress
// it gets a scope which is the scope of code that the err is came from
// it gets an error which is the error that just happened
// it gets a bool to know that the error is a serious one or not
// if isSerious is True then it will fatal the server after logging err
func logError(scope string, err error, isSerious bool) {

	go playBeep()

	if isSerious {
		log.Fatalln(scope, err)
	} else {
		fmt.Println(scope, err)
	}
}
