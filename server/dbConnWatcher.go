package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// processGate is the struct that we use to control whether incoming websocket connection should handle or not
// locker is the mutex that we use to lock isGateOpen to prevent race problems
// isGateOpen is the status of the processGate
// if isGateOpen is True then incoming websocket connections will be handled if False then no handling
type processGate struct {
	locker     sync.Mutex
	isGateOpen bool
}

// dbConnWatcher is a processGate method that watches our database connection
// if the database connection is dead and not responding then this will set the value of isGateOpen to False
// it gets a controller pointer for working with controller fields if necessary
func (gate *processGate) dbConnWatcher(c *controller) {

	for {
		time.Sleep(time.Second * 5)

		err := c.dbConn.ping()
		if err != nil && disconnectVerifier(c) {
			gate.locker.Lock()
			gate.isGateOpen = false
			gate.locker.Unlock()

			waitToEmptyOnlineClients(c)

			fmt.Println("Server shutdown due to database disconnection!")
			os.Exit(0)
		}
	}
}

// disconnectVerifier verifies database disconnection to make sure that it's for real
// it gets a controller pointer to work with our database connection
// it returns True is disconnection is for real and False if not
// here we test disconnection reality with three times of pinging
func disconnectVerifier(c *controller) bool {

	for i := 0; i < 3; i++ {
		err := c.dbConn.ping()
		if err == nil {
			return false
		}
	}

	return true
}

// waitToEmptyOnlineClients just waits to make sure that the onlineClients map is empty
// it checks the map every second gets the map from the given controller pointer
func waitToEmptyOnlineClients(c *controller) {

	for {
		time.Sleep(time.Millisecond)

		if c.onlineClientsLen() == 0 {
			return
		}
	}
}

// pGateCheck checks whether the process gate is open or not
// returns True if open and False if not
func (gate *processGate) pGateCheck() bool {

	gate.locker.Lock()
	defer gate.locker.Unlock()
	if !gate.isGateOpen {
		return false
	}

	return true
}
