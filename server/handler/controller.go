package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/mahditakrim/redFok/server/dataType"
	"github.com/mahditakrim/redFok/server/database"
	"github.com/mahditakrim/redFok/server/utility"
	"golang.org/x/net/websocket"
	"strings"
	"sync"
)

type Controller struct {
	mapLock       sync.Mutex
	onlineClients map[string]*websocket.Conn
	DbConn        database.DBHandler
}

func InitController(db database.DBHandler) *Controller {

	return &Controller{
		onlineClients: make(map[string]*websocket.Conn),
		DbConn:        db,
	}
}

func (c *Controller) addOnlineClient(conn *websocket.Conn, userName string) {

	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	c.onlineClients[userName] = conn
	go utility.Play()
	fmt.Println("online clients = ", len(c.onlineClients))
}

func (c *Controller) removeAndCloseOnlineClient(conn *websocket.Conn, userName string) {

	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	delete(c.onlineClients, userName)
	_ = conn.Close()
	fmt.Println("online clients = ", len(c.onlineClients))
}

func validateAuthentication(auth dataType.Authentication) bool {

	emptyHash := sha1.New()
	emptyHash.Write([]byte(""))
	emptyClientID := emptyHash.Sum(nil)

	if auth.UserName == "" || auth.ClientID == nil ||
		bytes.Compare(auth.ClientID, emptyClientID) == 0 {
		return false
	}

	if len(auth.ClientID) != sha1.Size || len(auth.UserName) > 50 {
		return false
	}

	if strings.Contains(auth.UserName, " ") {
		return false
	}

	return true
}

func (c *Controller) checkAuthentication(conn *websocket.Conn) string {

	var data []byte
	err := websocket.Message.Receive(conn, &data)
	if err != nil {
		utility.LogError("checkAuthentication-Receive", err, false)
		return ""
	}
	var auth dataType.Authentication
	err = json.Unmarshal(data, &auth)
	if err != nil {
		utility.LogError("checkAuthentication-Unmarshal", err, false)
		return ""
	}

	if !validateAuthentication(auth) {
		return ""
	}

	isExisted, err := c.DbConn.CheckClientID(auth.ClientID)
	if err != nil {
		utility.LogError("checkAuthentication-CheckClientID", err, false)
		return ""
	}
	if !isExisted {
		err := responseSender(conn, dataType.InvalidAuth)
		if err != nil {
			utility.LogError("checkAuthentication-isExisted", err, false)
		}

		return ""
	}

	result, err := c.DbConn.GetUserNameByClientID(auth.ClientID)
	if err != nil {
		utility.LogError("checkAuthentication-GetUserNameByClientID", err, false)
		return ""
	}
	if result != auth.UserName {
		err := responseSender(conn, dataType.InvalidAuth)
		if err != nil {
			utility.LogError("checkAuthentication-responseSender", err, false)
		}

		return ""
	}

	return result
}

func (c *Controller) checkIsClientOnline(userName string) bool {

	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	_, ok := c.onlineClients[userName]
	return ok
}

func responseSender(conn *websocket.Conn, flag string) error {

	err := websocket.JSON.Send(conn, dataType.Response{Value: flag})
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) OnlineClientsLen() int {

	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	return len(c.onlineClients)
}
