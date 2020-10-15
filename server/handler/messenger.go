package handler

import (
	"encoding/json"
	"github.com/mahditakrim/redFok/server/dataType"
	"github.com/mahditakrim/redFok/server/database"
	"github.com/mahditakrim/redFok/server/utility"
	"golang.org/x/net/websocket"
	"strings"
)

func (c *Controller) Messenger(conn *websocket.Conn) {

	userName := c.checkAuthentication(conn)
	if userName == "" || c.checkIsClientOnline(userName) {
		_ = conn.Close()
		return
	}

	c.addOnlineClient(conn, userName)

	err := responseSender(conn, dataType.Approve)
	if err != nil {
		utility.LogError("messenger-responseSender", err, false)
		c.removeAndCloseOnlineClient(conn, userName)
		return
	}

	ip := conn.Request().RemoteAddr[:strings.IndexByte(conn.Request().RemoteAddr, ':')]
	err = c.DbConn.ChangeIP(userName, ip)
	if err != nil {
		utility.LogError("messenger-ChangeIP", err, false)
		c.removeAndCloseOnlineClient(conn, userName)
		return
	}

	messages := c.checkUnseenMessages(userName, conn)
	if messages != nil {
		for _, message := range messages {
			err := c.DbConn.DeleteMessage("tbl_"+userName, message)
			if err != nil {
				utility.LogError("messenger-DeleteMessage", err, false)
				c.removeAndCloseOnlineClient(conn, userName)
				return
			}

			go func(message database.MessageData) {
				c.deliverMessage(conn, userName, dataType.ClientReceiveMessage{
					TimeStamp: message.TimeStamp,
					Text:      message.Text,
					Sender:    message.Sender,
				})
			}(message)
		}
	}

	c.runReceiver(conn, userName)
}

func (c *Controller) runReceiver(conn *websocket.Conn, userName string) {

	for {
		var data []byte
		err := websocket.Message.Receive(conn, &data)
		if err != nil {
			if c.checkIsClientOnline(userName) {
				utility.LogError("runReceiver-Receive", err, false)
				c.removeAndCloseOnlineClient(conn, userName)
			}
			return
		}
		var message dataType.ClientSendMessage
		err = json.Unmarshal(data, &message)
		if err != nil {
			utility.LogError("runReceiver-Unmarshal", err, false)
			c.removeAndCloseOnlineClient(conn, userName)
			return
		}

		go c.messageHandler(message, conn)
	}
}

func messageValidator(message *dataType.ClientSendMessage) bool {

	if message.Sender == "" || message.TimeStamp.String() == "" ||
		message.To == nil {
		return false
	}

	for _, user := range message.To {
		if user == "" {
			return false
		}
	}

	message.Text = strings.TrimSpace(message.Text)
	if message.Text == "" {
		return false
	}

	return true
}

func (c *Controller) messageHandler(message dataType.ClientSendMessage, conn *websocket.Conn) {

	if !messageValidator(&message) {
		return
	}

	for _, user := range message.To {
		if user == message.Sender {
			c.removeAndCloseOnlineClient(conn, message.Sender)
			return
		}
	}

	for _, user := range message.To {

		isClientExist, err := c.DbConn.CheckClientUserName(user)
		if err != nil {
			utility.LogError("messageHandler-CheckClientUserName", err, false)
			c.removeAndCloseOnlineClient(conn, message.Sender)
			return
		}
		if !isClientExist {
			if c.checkIsClientOnline(message.Sender) {
				err = responseSender(conn, dataType.NoSuchUser)
				if err != nil {
					utility.LogError("messageHandler-responseSender", err, false)
					c.removeAndCloseOnlineClient(conn, message.Sender)
				}
			}

			continue
		}

		go func(user string) {
			if c.checkIsClientOnline(user) {
				c.deliverMessage(c.onlineClients[user], user, dataType.ClientReceiveMessage{
					TimeStamp: message.TimeStamp,
					Text:      message.Text,
					Sender:    message.Sender,
				})

			} else {
				err := c.DbConn.InsertMessage("tbl_"+user, database.MessageData{
					TimeStamp: message.TimeStamp,
					Text:      message.Text,
					Sender:    message.Sender,
				})
				if err != nil {
					utility.LogError("messageHandler-InsertMessage", err, false)
					c.removeAndCloseOnlineClient(conn, message.Sender)
				}
			}
		}(user)

		if c.checkIsClientOnline(message.Sender) {
			err = responseSender(conn, dataType.Received)
			if err != nil {
				utility.LogError("messageHandler-responseSender", err, false)
				c.removeAndCloseOnlineClient(conn, message.Sender)
			}
		}
	}
}

func (c *Controller) deliverMessage(conn *websocket.Conn, userName string, message dataType.ClientReceiveMessage) {

	err := websocket.JSON.Send(conn, message)
	if err != nil {
		err := c.DbConn.InsertMessage("tbl_"+userName, database.MessageData{
			TimeStamp: message.TimeStamp,
			Text:      message.Text,
			Sender:    message.Sender,
		})
		if err != nil {
			utility.LogError("deliverMessage-InsertMessage", err, false)
		}

		c.removeAndCloseOnlineClient(conn, userName)

		utility.LogError("deliverMessage", err, false)
	}
}

func (c *Controller) checkUnseenMessages(userName string, conn *websocket.Conn) []database.MessageData {

	messages, err := c.DbConn.GetMessages("tbl_" + userName)
	if err != nil {
		utility.LogError("checkUnseenMessages", err, false)
		c.removeAndCloseOnlineClient(conn, userName)
		return nil
	}

	return messages
}
