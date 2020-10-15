package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/mahditakrim/redFok/server/dataType"
	"github.com/mahditakrim/redFok/server/database"
	"github.com/mahditakrim/redFok/server/utility"
	"golang.org/x/net/websocket"
	"strings"
)

func (c *Controller) Register(conn *websocket.Conn) {

	var data []byte
	err := websocket.Message.Receive(conn, &data)
	if err != nil {
		utility.LogError("register-Receive", err, false)
		_ = conn.Close()
		return
	}
	var reg dataType.Registration
	err = json.Unmarshal(data, &reg)
	if err != nil {
		utility.LogError("register-Unmarshal", err, false)
		_ = conn.Close()
		return
	}

	if !validateRegistration(reg) {
		_ = conn.Close()
		return
	}

	isClientExist, err := c.DbConn.CheckClientID(reg.ClientID)
	if err != nil {
		utility.LogError("register-CheckClientID", err, false)
		_ = conn.Close()
		return
	}
	if isClientExist {
		err := responseSender(conn, dataType.AlreadyReg)
		if err != nil {
			utility.LogError("register-responseSender", err, false)
		}

		_ = conn.Close()
		return
	}

	isUserNameInValid, err := c.DbConn.CheckClientUserName(reg.UserName)
	if err != nil {
		utility.LogError("register-CheckClientUserName", err, false)
		_ = conn.Close()
		return
	}
	if isUserNameInValid {
		err = responseSender(conn, dataType.InvalidUserName)
		if err != nil {
			utility.LogError("register-responseSender", err, false)
		}

		_ = conn.Close()
		return
	}

	//This db methods should be atomic, maybe by SProc
	//All the methods in a process should be atomic and reversible if some error occurred

	ip := conn.Request().RemoteAddr[:strings.IndexByte(conn.Request().RemoteAddr, ':')]
	err = c.DbConn.InsertUser(database.UserData{
		UserName: reg.UserName,
		ClientID: reg.ClientID,
		Name:     strings.TrimSpace(reg.Name),
		IP:       ip,
	})
	if err != nil {
		if err.(*mysql.MySQLError).Number == database.DuplicateErr {
			err = responseSender(conn, dataType.InvalidUserName)
			if err != nil {
				utility.LogError("register-isDuplicateError-responseSender", err, false)
			}

			_ = conn.Close()
			return
		}

		utility.LogError("register-InsertUser", err, false)
		_ = conn.Close()
		return
	}

	err = c.DbConn.CreateUserTable("tbl_" + reg.UserName)
	if err != nil {
		utility.LogError("register-CreateUserTable", err, false)
		_ = conn.Close()
		return
	}

	err = responseSender(conn, dataType.Approve)
	if err != nil {
		utility.LogError("register-responseSender", err, false)
		_ = conn.Close()
		return
	}

	go utility.Play()
	fmt.Println(reg.UserName + " registered")
	_ = conn.Close()
}

func validateRegistration(reg dataType.Registration) bool {

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
