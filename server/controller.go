package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"strings"
	"sync"
)

// onlineClient is the struct that we use to keep online clients map with its mutex together.
// mapLock is the mutex that we use to lock onlineClients map to prevent race problems.
// clients is the map we use to store online users's websocket connection and key of the map is client's userName.
type onlineClient struct {
	mapLock sync.Mutex
	clients map[string]*websocket.Conn
}

// controller is the struct that we use to init our server for keeping online clients and the database connection.
// onlineClients is the struct that we use to keep online clients map with its mutex together.
// dbConn is the database connection holder we use to communicate with mysql database.
type controller struct {
	onlineClients onlineClient
	dbConn        dbHandler
}

// initNewController inits a controller and returns it as pointer.
// it gets a database connection which is in type of dbHandler struct.
func initNewController(db dbHandler) *controller {

	return &controller{
		onlineClients: onlineClient{clients: make(map[string]*websocket.Conn)},
		dbConn:        db,
	}
}

// addOnlineClient is a controller method that adds a new user client to the onlineClients map.
// it gets user's websocket connection and the userName as the key for the map.
func (c *controller) addOnlineClient(conn *websocket.Conn, userName string) {

	c.onlineClients.mapLock.Lock()
	defer c.onlineClients.mapLock.Unlock()
	c.onlineClients.clients[userName] = conn
	go playBeep()
	fmt.Println("online clients = ", len(c.onlineClients.clients))
}

// removeAndCloseOnlineClient is a controller method that removes and also closes the client from onlineClients map.
// it tries to close the websocket connection whether its already close or not.
// it gets a websocket connection for closing and the userName as the key for delete.
func (c *controller) removeAndCloseOnlineClient(conn *websocket.Conn, userName string) {

	c.onlineClients.mapLock.Lock()
	defer c.onlineClients.mapLock.Unlock()
	delete(c.onlineClients.clients, userName)
	_ = conn.Close()
	fmt.Println("online clients = ", len(c.onlineClients.clients))
}

// validateAuthentication validates an authentication in terms of data appearance.
// it gets an authentication struct.
// it returns True if all checks wend alright and False if not.
func validateAuthentication(auth authentication) bool {

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

// checkAuthentication is a controller method that checks whether a client is allowed to communicate with the server or not.
// it gets a websocket connection as the incoming client.
// it returns the client's userName if authentication went alright and returns an empty string if not.
func (c *controller) checkAuthentication(conn *websocket.Conn) string {

	var data []byte
	err := websocket.Message.Receive(conn, &data)
	if err != nil {
		logError("checkAuthentication-Receive", err)
		return ""
	}
	var auth authentication
	err = json.Unmarshal(data, &auth)
	if err != nil {
		logError("checkAuthentication-Unmarshal", err)
		return ""
	}

	if !validateAuthentication(auth) {
		return ""
	}

	isExisted, err := c.dbConn.checkClientID(auth.ClientID)
	if err != nil {
		logError("checkAuthentication-checkClientID", err)
		return ""
	}
	if !isExisted {
		err := responseSender(conn, invalidAuth)
		if err != nil {
			logError("checkAuthentication-isExisted", err)
		}

		return ""
	}

	result, err := c.dbConn.getUserNameByClientID(auth.ClientID)
	if err != nil {
		logError("checkAuthentication-getUserNameByClientID", err)
		return ""
	}
	if result != auth.UserName {
		err := responseSender(conn, invalidAuth)
		if err != nil {
			logError("checkAuthentication-responseSender", err)
		}

		return ""
	}

	return result
}

// checkIsClientOnline is a controller method that checks whether a client is in onlineClients map or not.
// it gets client's userName as the map key.
// it returns True if the userName is in map and returns False if not.
func (c *controller) checkIsClientOnline(userName string) bool {

	c.onlineClients.mapLock.Lock()
	defer c.onlineClients.mapLock.Unlock()
	_, ok := c.onlineClients.clients[userName]
	return ok
}

// responseSender sends server responses to the clients with the specific response flag.
// it gets a websocket connection for sending and the flag as the response flag.
// returns error if something went wrong.
func responseSender(conn *websocket.Conn, flag string) error {

	err := websocket.JSON.Send(conn, response{Value: flag})
	if err != nil {
		return err
	}

	return nil
}

// onlineClientsLen is a controller method that return the onlineClients map's length.
// it returns length as an int.
// the returned value is the map's length at the method's run time.
// and because of using mutex the returned value is accurate.
func (c *controller) onlineClientsLen() int {

	c.onlineClients.mapLock.Lock()
	defer c.onlineClients.mapLock.Unlock()
	return len(c.onlineClients.clients)
}
