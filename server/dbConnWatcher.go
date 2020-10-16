package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type processGate struct {
	locker     sync.Mutex
	isGateOpen bool
}

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

func disconnectVerifier(c *controller) bool {

	for i := 0; i < 3; i++ {
		err := c.dbConn.ping()
		if err == nil {
			return false
		}
	}

	return true
}

func waitToEmptyOnlineClients(c *controller) {

	for {
		time.Sleep(time.Millisecond)

		if c.onlineClientsLen() == 0 {
			return
		}
	}
}

func (gate *processGate) pGateCheck() bool {

	gate.locker.Lock()
	defer gate.locker.Unlock()
	if !gate.isGateOpen {
		return false
	}

	return true
}
