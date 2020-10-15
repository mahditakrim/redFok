package utility

import (
	"fmt"
	"log"
)

//We may want logging processes here . . .

func LogError(scope string, err error, isSerious bool) {

	go Play()

	if isSerious {
		log.Fatalln(scope, err)
	} else {
		fmt.Println(scope, err)
	}
}
