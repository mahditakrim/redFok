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

// Controller is the struct that we use to init our server for keeping online clients and the database connection
// mapLock is the mutex we use to lock onlineClients map to prevent race problems
// onlineClients is the map we use to store online users's websocket connection and key of the map is client's userName
// DbConn is the database connection holder we use to communicate with mysql database
type Controller struct {
	mapLock       sync.Mutex
	onlineClients map[string]*websocket.Conn
	DbConn        database.DBHandler
}

// InitNewController inits a Controller and returns it as pointer
// it gets a database connection which is in type of DBHandler struct
func InitNewController(db database.DBHandler) *Controller {

	return &Controller{
		onlineClients: make(map[string]*websocket.Conn),
		DbConn:        db,
	}
}

// addOnlineClient is a controller method that adds a new user client to the onlineClients map
// it gets user's websocket connection and the userName as the key for the map
func (c *Controller) addOnlineClient(conn *websocket.Conn, userName string) {

	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	c.onlineClients[userName] = conn
	go utility.Play()
	fmt.Println("online clients = ", len(c.onlineClients))
}

// removeAndCloseOnlineClient is a Controller method that removes and also closes the client from onlineClients map
// it tries to close the websocket connection whether its already close or not
// it gets a websocket connection for closing and the userName as the key for delete
func (c *Controller) removeAndCloseOnlineClient(conn *websocket.Conn, userName string) {

	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	delete(c.onlineClients, userName)
	_ = conn.Close()
	fmt.Println("online clients = ", len(c.onlineClients))
}

// validateAuthentication validates an Authentication in terms of data appearance
// it gets an Authentication struct
// it returns True if all checks wend alright and False if not
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

// checkAuthentication is a Controller method that checks whether a client is allowed to communicate with the server or not
// it gets a websocket connection as the incoming client
// it returns the client's userName if authentication went alright and returns an empty string if not
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

// checkIsClientOnline is a Controller method that checks whether a client is in onlineClients map or not
// it gets client's userName as the map key
// it returns True if the userName is in map and returns False if not
func (c *Controller) checkIsClientOnline(userName string) bool {

	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	_, ok := c.onlineClients[userName]
	return ok
}

// responseSender sends server responses to the clients with the specific response flag
// it gets a websocket connection for sending and the flag as the response flag
// returns error if something went wrong
func responseSender(conn *websocket.Conn, flag string) error {

	err := websocket.JSON.Send(conn, dataType.Response{Value: flag})
	if err != nil {
		return err
	}

	return nil
}

// OnlineClientsLen is a Controller method that return the onlineClients map's length
// it returns length as an int
// the returned value is the map's length at the method's run time
// and because of using mutex the returned value is accurate
func (c *Controller) OnlineClientsLen() int {

	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	return len(c.onlineClients)
}
