package main

import (
	"encoding/json"
	"golang.org/x/net/websocket"
	"strings"
)

// messenger is a controller pointer method that handles messaging process
// it gets a websocket connection pointer and uses it as incoming user
func (c *controller) messenger(conn *websocket.Conn) {

	userName := c.checkAuthentication(conn)
	if userName == "" || c.checkIsClientOnline(userName) {
		_ = conn.Close()
		return
	}

	c.addOnlineClient(conn, userName)

	err := responseSender(conn, approve)
	if err != nil {
		logError("messenger-responseSender", err, false)
		c.removeAndCloseOnlineClient(conn, userName)
		return
	}

	ip := conn.Request().RemoteAddr[:strings.IndexByte(conn.Request().RemoteAddr, ':')]
	err = c.dbConn.changeIP(userName, ip)
	if err != nil {
		logError("messenger-changeIP", err, false)
		c.removeAndCloseOnlineClient(conn, userName)
		return
	}

	messages := c.checkUnseenMessages(userName, conn)
	if messages != nil {
		for _, message := range messages {
			err := c.dbConn.deleteMessage("tbl_"+userName, message)
			if err != nil {
				logError("messenger-deleteMessage", err, false)
				c.removeAndCloseOnlineClient(conn, userName)
				return
			}

			go func(message messageData) {
				c.deliverMessage(conn, userName, clientReceiveMessage{
					TimeStamp: message.timeStamp,
					Text:      message.text,
					Sender:    message.sender,
				})
			}(message)
		}
	}

	c.runReceiver(conn, userName)
}

// runReceiver runs a websocket Receiver on the given conn
// it gets a websocket connection pointer to listen and receive
// it gets a userName which is the clients authorized userName
func (c *controller) runReceiver(conn *websocket.Conn, userName string) {

	for {
		var data []byte
		err := websocket.Message.Receive(conn, &data)
		if err != nil {
			if c.checkIsClientOnline(userName) {
				logError("runReceiver-Receive", err, false)
				c.removeAndCloseOnlineClient(conn, userName)
			}
			return
		}
		var message clientSendMessage
		err = json.Unmarshal(data, &message)
		if err != nil {
			logError("runReceiver-Unmarshal", err, false)
			c.removeAndCloseOnlineClient(conn, userName)
			return
		}

		go c.messageHandler(message, conn)
	}
}

// messageValidator validates a clientSendMessage in terms of data appearance
// it gets a clientSendMessage pointer for space trimming so the value will be change globally
// it returns True if everything wend alright and False if not
func messageValidator(message *clientSendMessage) bool {

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

// messageHandler is a controller pointer method that handles every single clientSendMessage that runReceiver receives
// it gets a clientSendMessage fro processing
// it gets a websocket connection pointer as the user that has sent the message
func (c *controller) messageHandler(message clientSendMessage, conn *websocket.Conn) {

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

		isClientExist, err := c.dbConn.checkClientUserName(user)
		if err != nil {
			logError("messageHandler-checkClientUserName", err, false)
			c.removeAndCloseOnlineClient(conn, message.Sender)
			return
		}
		if !isClientExist {
			if c.checkIsClientOnline(message.Sender) {
				err = responseSender(conn, noSuchUser)
				if err != nil {
					logError("messageHandler-responseSender", err, false)
					c.removeAndCloseOnlineClient(conn, message.Sender)
				}
			}

			continue
		}

		go func(user string) {
			if c.checkIsClientOnline(user) {
				c.deliverMessage(c.onlineClients[user], user, clientReceiveMessage{
					TimeStamp: message.TimeStamp,
					Text:      message.Text,
					Sender:    message.Sender,
				})

			} else {
				err := c.dbConn.insertMessage("tbl_"+user, messageData{
					timeStamp: message.TimeStamp,
					text:      message.Text,
					sender:    message.Sender,
				})
				if err != nil {
					logError("messageHandler-insertMessage", err, false)
					c.removeAndCloseOnlineClient(conn, message.Sender)
				}
			}
		}(user)

		if c.checkIsClientOnline(message.Sender) {
			err = responseSender(conn, received)
			if err != nil {
				logError("messageHandler-responseSender", err, false)
				c.removeAndCloseOnlineClient(conn, message.Sender)
			}
		}
	}
}

// deliverMessage is a controller pointer method that delivers a clientReceiveMessage to the userName
// it gets a websocket connection pointer and a userName as the users info for sending the message to
func (c *controller) deliverMessage(conn *websocket.Conn, userName string, message clientReceiveMessage) {

	err := websocket.JSON.Send(conn, message)
	if err != nil {
		err := c.dbConn.insertMessage("tbl_"+userName, messageData{
			timeStamp: message.TimeStamp,
			text:      message.Text,
			sender:    message.Sender,
		})
		if err != nil {
			logError("deliverMessage-insertMessage", err, false)
		}

		c.removeAndCloseOnlineClient(conn, userName)

		logError("deliverMessage", err, false)
	}
}

// checkUnseenMessages is a controller pointer method that checks whether userName has unseen messages
// it gets a websocket connection pointer and a userName as user info to check
// it returns a slice of messageData as the unseen messages
// returns nil if there is no unseen message in the database
func (c *controller) checkUnseenMessages(userName string, conn *websocket.Conn) []messageData {

	messages, err := c.dbConn.getMessages("tbl_" + userName)
	if err != nil {
		logError("checkUnseenMessages", err, false)
		c.removeAndCloseOnlineClient(conn, userName)
		return nil
	}

	return messages
}
