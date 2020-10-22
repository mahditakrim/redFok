package main

import (
	"fmt"
	"golang.org/x/net/websocket"
)

// deleter is a controller pointer method that handles user deletion process.
// it gets a websocket connection pinter and uses it as the user connection that will be deleted.
func (c *controller) deleter(conn *websocket.Conn) {

	userName := c.checkAuthentication(conn)
	if userName == "" {
		_ = conn.Close()
		return
	}

	err := c.dbConn.deleteUserAndTable(userName)
	if err != nil {
		logError("deleter-deleteUserAndTable", err)
		c.removeAndCloseOnlineClient(userName)
		return
	}

	err = responseSender(conn, approved)
	if err != nil {
		logError("deleter-responseSender", err)
	}

	c.removeAndCloseOnlineClient(userName)

	go playBeep()
	fmt.Println(userName, " deleted")
}
