package main

import (
	"fmt"
	"golang.org/x/net/websocket"
)

// deleter is a controller pointer method that handles user deletion
// it gets a websocket connection pinter and uses it as the user connection that will be deleted
func (c *controller) deleter(conn *websocket.Conn) {

	userName := c.checkAuthentication(conn)
	if userName == "" {
		_ = conn.Close()
		return
	}

	err := c.dbConn.deleteUserAndTable(userName)
	if err != nil {
		logError("deleter-deleteUserAndTable", err, false)
		c.removeAndCloseOnlineClient(conn, userName)
		return
	}

	err = responseSender(conn, approve)
	if err != nil {
		logError("deleter-responseSender", err, false)
	}

	c.removeAndCloseOnlineClient(conn, userName)

	go playBeep()
	fmt.Println(userName, " deleted")
}
