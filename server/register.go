package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/net/websocket"
	"strings"
)

// register is a controller pointer method that handles registration process.
// it gets a websocket connection pointer as the incoming user who wants to register.
func (c *controller) register(conn *websocket.Conn) {

	defer func() {
		if r := recover(); r != nil {
			logError(r.(errScope).scope, r.(errScope).err)
			_ = conn.Close()
		}
	}()

	var data []byte
	err := websocket.Message.Receive(conn, &data)
	if err != nil {
		panic(errScope{scope: "register-Receive", err: err})
	}
	var reg registration
	err = json.Unmarshal(data, &reg)
	if err != nil {
		panic(errScope{scope: "register-Unmarshal", err: err})
	}

	if !validateRegistration(reg) {
		_ = conn.Close()
		return
	}

	isClientExist, err := c.dbConn.checkClientID(reg.ClientID)
	if err != nil {
		panic(errScope{scope: "register-checkClientID", err: err})
	}
	if isClientExist {
		err = responseSender(conn, alreadyReg)
		if err != nil {
			panic(errScope{scope: "register-responseSender", err: err})
		}

		_ = conn.Close()
		return
	}

	isUserNameInValid, err := c.dbConn.checkClientUserName(reg.UserName)
	if err != nil {
		panic(errScope{scope: "register-checkClientUserName", err: err})
	}
	if isUserNameInValid {
		err = responseSender(conn, invalidUserName)
		if err != nil {
			panic(errScope{scope: "register-invalid-responseSender", err: err})
		}

		_ = conn.Close()
		return
	}

	userIP := conn.Request().RemoteAddr[:strings.IndexByte(conn.Request().RemoteAddr, ':')]
	err = c.dbConn.insertUserAndCreateTable(userData{
		userName: reg.UserName,
		clientID: reg.ClientID,
		name:     strings.TrimSpace(reg.Name),
		ip:       userIP,
	}, "tbl_"+reg.UserName)
	if err != nil {
		if err.(*mysql.MySQLError).Number == duplicateErr {
			err = responseSender(conn, invalidUserName)
			if err != nil {
				panic(errScope{scope: "register-duplicate-responseSender", err: err})
			}

			_ = conn.Close()
			return
		}

		panic(errScope{scope: "register-insertUserAndCreateTable", err: err})
	}

	err = responseSender(conn, approved)
	if err != nil {
		panic(errScope{scope: "register-approved-responseSender", err: err})
	}

	go playBeep()
	fmt.Println(reg.UserName + " registered")
	_ = conn.Close()
}

// validateRegistration validates a registration in terms of data appearance.
// it gets a registration to validate.
// it returns True if everything went alright and False if not.
func validateRegistration(reg registration) bool {

	emptyHash := sha1.New()
	emptyHash.Write([]byte(""))
	emptyClientID := emptyHash.Sum(nil)

	if strings.TrimSpace(reg.UserName) == "" ||
		reg.ClientID == nil ||
		bytes.Compare(reg.ClientID, emptyClientID) == 0 {
		return false
	}

	if len(reg.ClientID) != sha1.Size ||
		len(reg.Name) > 50 ||
		len(reg.UserName) > 50 {
		return false
	}

	if strings.Contains(reg.UserName, " ") {
		return false
	}

	return true
}
