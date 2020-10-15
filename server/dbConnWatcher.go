package main

import (
	"fmt"
	"github.com/mahditakrim/redFok/server/handler"
	"os"
	"sync"
	"time"
)

type processGate struct {
	locker     sync.Mutex
	isGateOpen bool
}

func (gate *processGate) dbConnWatcher(c *handler.Controller) {

	for {
		time.Sleep(time.Second * 5)

		err := c.DbConn.Ping()
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

func disconnectVerifier(c *handler.Controller) bool {

	for i := 0; i < 3; i++ {
		err := c.DbConn.Ping()
		if err == nil {
			return false
		}
	}

	return true
}

func waitToEmptyOnlineClients(c *handler.Controller) {

	for {
		time.Sleep(time.Millisecond)

		if c.OnlineClientsLen() == 0 {
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
