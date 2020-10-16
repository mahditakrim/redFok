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

func (c *controller) register(conn *websocket.Conn) {

	var data []byte
	err := websocket.Message.Receive(conn, &data)
	if err != nil {
		logError("register-Receive", err, false)
		_ = conn.Close()
		return
	}
	var reg registration
	err = json.Unmarshal(data, &reg)
	if err != nil {
		logError("register-Unmarshal", err, false)
		_ = conn.Close()
		return
	}

	if !validateRegistration(reg) {
		_ = conn.Close()
		return
	}

	isClientExist, err := c.dbConn.checkClientID(reg.ClientID)
	if err != nil {
		logError("register-checkClientID", err, false)
		_ = conn.Close()
		return
	}
	if isClientExist {
		err := responseSender(conn, alreadyReg)
		if err != nil {
			logError("register-responseSender", err, false)
		}

		_ = conn.Close()
		return
	}

	isUserNameInValid, err := c.dbConn.checkClientUserName(reg.UserName)
	if err != nil {
		logError("register-checkClientUserName", err, false)
		_ = conn.Close()
		return
	}
	if isUserNameInValid {
		err = responseSender(conn, invalidUserName)
		if err != nil {
			logError("register-responseSender", err, false)
		}

		_ = conn.Close()
		return
	}

	//This db methods should be atomic, maybe by SProc
	//All the methods in a process should be atomic and reversible if some error occurred

	ip := conn.Request().RemoteAddr[:strings.IndexByte(conn.Request().RemoteAddr, ':')]
	err = c.dbConn.insertUser(userData{
		userName: reg.UserName,
		clientID: reg.ClientID,
		name:     strings.TrimSpace(reg.Name),
		ip:       ip,
	})
	if err != nil {
		if err.(*mysql.MySQLError).Number == duplicateErr {
			err = responseSender(conn, invalidUserName)
			if err != nil {
				logError("register-isDuplicateError-responseSender", err, false)
			}

			_ = conn.Close()
			return
		}

		logError("register-insertUser", err, false)
		_ = conn.Close()
		return
	}

	err = c.dbConn.createUserTable("tbl_" + reg.UserName)
	if err != nil {
		logError("register-createUserTable", err, false)
		_ = conn.Close()
		return
	}

	err = responseSender(conn, approve)
	if err != nil {
		logError("register-responseSender", err, false)
		_ = conn.Close()
		return
	}

	go playBeep()
	fmt.Println(reg.UserName + " registered")
	_ = conn.Close()
}

func validateRegistration(reg registration) bool {

	emptyHash := sha1.New()
	emptyHash.Write([]byte(""))
	emptyClientID := emptyHash.Sum(nil)

	if strings.TrimSpace(reg.Name) == "" ||
		strings.TrimSpace(reg.UserName) == "" ||
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