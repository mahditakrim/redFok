package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

// duplicateErr is the error code of mysql for inserting duplicate value to the database
// this error can be produced when you have 'UNIQUE' fields
const duplicateErr = 1062

// dbHandler contains the database connection that we use in the whole project
type dbHandler struct {
	db *sql.DB
}

// messageData is the struct that we use to insert messages into database
// timeStamp is the time that user has sent the message
// text is user's text message
// sender is the user's 'userName' that has sent the message
type messageData struct {
	timeStamp time.Time
	text      string
	sender    string
}

// userData is the struct that we use to insert new user's data into database
// userName is the unique identifier that we use to detect different users from each other
// ClientID is the unique identifier that we use for 2FA and ...
// name is the optional name that user can choose for profile
// ip is the user's connection ip and it's for tracking user's connection and filtering stuffs
type userData struct {
	userName string
	clientID []byte
	name     string
	ip       string
}

// createDBConnection inits a new mysql connection and returns it via a dbHandler pointer
// And returns error if the init went wrong
// it also pings the mysql server to ensure that the server is alive and responding
// connString is connection string to the mysql server
func createDBConnection(connString string) (*dbHandler, error) {

	dbConn, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	if dbConn.Ping() != nil {
		return nil, nil
	}

	return &dbHandler{db: dbConn}, nil
}

// getUserNameByClientID gets an ID in byte form and returns the userName of that ID
// if something went wrong, returns the error
func (dbConn dbHandler) getUserNameByClientID(ID []byte) (string, error) {

	row := dbConn.db.QueryRow(
		"SELECT userName FROM tbl_users WHERE ClientID = ?", ID)
	var username string
	err := row.Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

// checkClientID checks whether the ID exists in the database or not
// returns True if exists and False if not, error if something went wrong
func (dbConn dbHandler) checkClientID(ID []byte) (bool, error) {

	row := dbConn.db.QueryRow(
		"SELECT EXISTS (SELECT * FROM tbl_users WHERE ClientID = ?)", ID)
	var result bool
	err := row.Scan(&result)
	if err != nil {
		return false, err
	}

	return result, nil
}

// checkClientUserName checks whether the userName exists in the database or not
// returns True if exists and False if not, error if something went wrong
func (dbConn dbHandler) checkClientUserName(userName string) (bool, error) {

	row := dbConn.db.QueryRow(
		"SELECT EXISTS (SELECT * FROM tbl_users WHERE userName = ?)", userName)
	var result bool
	err := row.Scan(&result)
	if err != nil {
		return false, err
	}

	return result, nil
}

// insertMessage inserts a messageData into the given table
// returns error if something went wrong
func (dbConn dbHandler) insertMessage(table string, message messageData) error {

	_, err := dbConn.db.Exec("INSERT INTO "+table+" VALUE (?, ?, ?)",
		message.timeStamp, message.text, message.sender)
	if err != nil {
		return err
	}

	return nil
}

// getMessages gets all the messageData from the given table and returns them as a slice
// returns error if something went wrong
func (dbConn dbHandler) getMessages(table string) ([]messageData, error) {

	rows, err := dbConn.db.Query("SELECT * FROM " + table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []messageData

	for rows.Next() {
		var data messageData
		err := rows.Scan(&data.timeStamp, &data.text, &data.sender)
		if err != nil {
			return nil, err
		}
		result = append(result, data)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// deleteMessage deletes given message from the given table
// returns error if something went wrong
func (dbConn dbHandler) deleteMessage(table string, message messageData) error {

	_, err := dbConn.db.Exec("DELETE FROM "+
		table+" WHERE timeStamp = ? AND text = ? AND sender = ?",
		message.timeStamp, message.text, message.sender)
	if err != nil {
		return err
	}

	return nil
}

// insertUserAndCreateTable inserts a userData into the database and also creates a table with the given table name
// returns error if something went wrong
func (dbConn dbHandler) insertUserAndCreateTable(user userData, table string) error {

	tx, err := dbConn.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO tbl_users VALUE (?, ?, ?, ?)",
		user.userName, user.clientID, user.name, user.ip)
	if err != nil {
		return err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS " + table +
		" (timeStamp DATETIME NOT NULL," +
		" text TEXT NOT NULL," +
		" sender VARCHAR(50) NOT NULL)")
	if err != nil {
		_ = tx.Rollback() //this will return error so we have to handle it
		return err
	}

	return tx.Commit()
}

// changeIP changes the ip of the given userName by the given ip
// returns error if something went wrong
func (dbConn dbHandler) changeIP(userName string, ip string) error {

	_, err := dbConn.db.Exec("UPDATE tbl_users SET ip = ? WHERE userName = ?",
		ip, userName)
	if err != nil {
		return err
	}

	return nil
}

// deleteUserAndTable deletes both user record from tbl_users and the table of user
// this does it with two execution one by one
// this should be an atomic job and do both executions at once
// returns error if one of the executions went wrong
func (dbConn dbHandler) deleteUserAndTable(user string) error {

	tx, err := dbConn.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM tbl_users WHERE userName = ?", user)
	if err != nil {
		return err
	}

	table := "tbl_" + user
	_, err = tx.Exec("DROP TABLE " + table)
	if err != nil {
		_ = tx.Rollback() //this will return error so we have to handle it
		return err
	}

	return tx.Commit()
}

// ping simply pings the mysql service provider and returns error if no answer
// if we got error then it means that mysql is not alive and responding
func (dbConn dbHandler) ping() error {

	err := dbConn.db.Ping()
	if err != nil {
		return err
	}

	return nil
}
