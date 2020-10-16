package main

import (
	"fmt"
	"golang.org/x/net/websocket"
)

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
