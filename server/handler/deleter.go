package handler

import (
	"fmt"
	"github.com/mahditakrim/redFok/server/dataType"
	"github.com/mahditakrim/redFok/server/utility"
	"golang.org/x/net/websocket"
)

func (c *Controller) Deleter(conn *websocket.Conn) {

	userName := c.checkAuthentication(conn)
	if userName == "" {
		_ = conn.Close()
		return
	}

	err := c.DbConn.DeleteUserAndTable(userName)
	if err != nil {
		utility.LogError("deleter-DeleteUserAndTable", err, false)
		c.removeAndCloseOnlineClient(conn, userName)
		return
	}

	err = responseSender(conn, dataType.Approve)
	if err != nil {
		utility.LogError("deleter-responseSender", err, false)
	}

	c.removeAndCloseOnlineClient(conn, userName)

	go utility.Play()
	fmt.Println(userName, " deleted")
}
