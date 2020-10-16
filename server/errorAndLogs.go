package main

import (
	"fmt"
	"log"
)

//We may want logging processes here . . .

func logError(scope string, err error, isSerious bool) {

	go playBeep()

	if isSerious {
		log.Fatalln(scope, err)
	} else {
		fmt.Println(scope, err)
	}
}
