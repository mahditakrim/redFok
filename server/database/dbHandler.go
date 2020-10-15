package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

// DuplicateErr is the error code of mysql for inserting duplicate value to the database
// this error can be produced when you have 'UNIQUE' fields
const DuplicateErr = 1062

// DBHandler contains the database connection that we use in the whole project
type DBHandler struct {
	db *sql.DB
}

// MessageData is the struct that we use to insert messages into database
// TimeStamp is the time that user has sent the message
// Text is user's text message
// Sender is the user's 'userName' that has sent the message
type MessageData struct {
	TimeStamp time.Time
	Text      string
	Sender    string
}

// UserData is the struct that we use to insert new user's data into database
// UserName is the unique identifier that we use to detect different users from each other
// ClientID is the unique identifier that we use for 2FA and ...
// Name is the optional name that user can choose for profile
// IP is the user's connection ip and it's for tracking user's connection and filtering stuffs
type UserData struct {
	UserName string
	ClientID []byte
	Name     string
	IP       string
}

// CreateDBConnection inits a new mysql connection and returns it via a DBHandler pointer
// And returns error if the init went wrong
// it also pings the mysql server to ensure that the server is alive and responding
// connString is connection string to the mysql server
func CreateDBConnection(connString string) (*DBHandler, error) {

	dbConn, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	if dbConn.Ping() != nil {
		return nil, nil
	}

	return &DBHandler{db: dbConn}, nil
}

// GetUserNameByClientID gets an ID in byte form and returns the userName of that ID
// if something went wrong, returns the error
func (dbConn DBHandler) GetUserNameByClientID(ID []byte) (string, error) {

	row := dbConn.db.QueryRow(
		"SELECT userName FROM tbl_users WHERE clientID = ?", ID)
	var username string
	err := row.Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

// CheckClientID checks whether the ID exists in the database or not
// returns True if exists and False if not, error if something went wrong
func (dbConn DBHandler) CheckClientID(ID []byte) (bool, error) {

	row := dbConn.db.QueryRow(
		"SELECT EXISTS (SELECT * FROM tbl_users WHERE clientID = ?)", ID)
	var result bool
	err := row.Scan(&result)
	if err != nil {
		return false, err
	}

	return result, nil
}

// CheckClientUserName checks whether the userName exists in the database or not
// returns True if exists and False if not, error if something went wrong
func (dbConn DBHandler) CheckClientUserName(userName string) (bool, error) {

	row := dbConn.db.QueryRow(
		"SELECT EXISTS (SELECT * FROM tbl_users WHERE userName = ?)", userName)
	var result bool
	err := row.Scan(&result)
	if err != nil {
		return false, err
	}

	return result, nil
}

// InsertMessage inserts a messageData into the given table
// returns error if something went wrong
func (dbConn DBHandler) InsertMessage(table string, message MessageData) error {

	_, err := dbConn.db.Exec("INSERT INTO "+table+" VALUE (?, ?, ?)",
		message.TimeStamp, message.Text, message.Sender)
	if err != nil {
		return err
	}

	return nil
}

// GetMessages gets all the messageData from the given table and returns them as a slice
// returns error if something went wrong
func (dbConn DBHandler) GetMessages(table string) ([]MessageData, error) {

	rows, err := dbConn.db.Query("SELECT * FROM " + table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []MessageData

	for rows.Next() {
		var data MessageData
		err := rows.Scan(&data.TimeStamp, &data.Text, &data.Sender)
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

// DeleteMessage deletes given message from the given table
// returns error if something went wrong
func (dbConn DBHandler) DeleteMessage(table string, message MessageData) error {

	_, err := dbConn.db.Exec("DELETE FROM "+
		table+" WHERE timeStamp = ? AND text = ? AND sender = ?",
		message.TimeStamp, message.Text, message.Sender)
	if err != nil {
		return err
	}

	return nil
}

// InsertUser inserts a userData into the database
// returns error if something went wrong
func (dbConn DBHandler) InsertUser(user UserData) error {

	_, err := dbConn.db.Exec("INSERT INTO tbl_users VALUE (?, ?, ?, ?)",
		user.UserName, user.ClientID, user.Name, user.IP)
	if err != nil {
		return err
	}

	return nil
}

// ChangeIP changes the IP of the given userName by the given ip
// returns error if something went wrong
func (dbConn DBHandler) ChangeIP(userName string, ip string) error {

	_, err := dbConn.db.Exec("UPDATE tbl_users SET IP = ? WHERE userName = ?",
		ip, userName)
	if err != nil {
		return err
	}

	return nil
}

// CreateUserTable creates a table with the given table name
// if the given table is already existing then nothing happens
// returns error if something went wrong
func (dbConn DBHandler) CreateUserTable(table string) error {

	_, err := dbConn.db.Exec("CREATE TABLE IF NOT EXISTS " + table +
		" (timeStamp DATETIME NOT NULL," +
		" text TEXT NOT NULL," +
		" sender VARCHAR(50) NOT NULL)")
	if err != nil {
		return err
	}

	return nil
}

// DeleteUserAndTable deletes both user record from tbl_users and the table of user
// this does it with two execution one by one
// this should be an atomic job and do both executions at once
// returns error if one of the executions went wrong
func (dbConn DBHandler) DeleteUserAndTable(user string) error {

	//this should be atomic

	table := "tbl_" + user
	_, err := dbConn.db.Exec("DELETE FROM tbl_users WHERE userName = ?", user)
	if err != nil {
		return err
	}

	_, err = dbConn.db.Exec("DROP TABLE " + table)
	if err != nil {
		return err
	}

	return nil
}

// Ping simply pings the mysql service provider and returns error if no answer
// if we got error then it means that mysql is not alive and responding
func (dbConn DBHandler) Ping() error {

	err := dbConn.db.Ping()
	if err != nil {
		return err
	}

	return nil
}
