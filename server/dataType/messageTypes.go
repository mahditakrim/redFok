package dataType

import "time"

// these are the flags that we use to send server responses
// the values contains three digits of the flag name
// Received is the flag that server uses to response to clients saying that we processed your message successfully
// Approve is the flag that server uses to response to clients saying that your operation processed successfully
// InvalidUserName is the flag that server uses to response to clients saying that username is not valid to be save in database
// NoSuchUser is the flag that server uses to response to clients saying that the username you want to send message to it, is not existing in database
// AlreadyReg is the flag that server uses to response to clients saying that the clientID is already existing in database
// InvalidAuth is the flag that server uses to response to clients saying that Authentication is not valid
const (
	Received        = "RCV"
	Approve         = "APV"
	InvalidUserName = "IUN"
	NoSuchUser      = "NSU"
	AlreadyReg      = "ART"
	InvalidAuth     = "IAT"
)

// Response is the json struct that we use to send server responses
// Value can be either one of the above flags
type Response struct {
	Value string `json:"value"`
}

// Authentication is the json struct that clients should send at the very beginning of connection
// ClientID is the unique identifier that we use for 2FA and ...
// UserName is the unique identifier that we use to detect different users from each other
type Authentication struct {
	ClientID []byte `json:"clientID"`
	UserName string `json:"userName"`
}

// Registration is the json struct that clients should use for sending info to register an account
// ClientID is the unique identifier that we use for 2FA and ...
// UserName is the unique identifier that we use to detect different users from each other
// Name is the optional name that user can choose for profile
type Registration struct {
	ClientID []byte `json:"clientID"`
	UserName string `json:"userName"`
	Name     string `json:"name"`
}

// ClientSendMessage is the json struct that clients should use for sending their messages
// server processes client messages in this json format
// TimeStamp is the time that user has sent the message
// Text is user's text message
// Sender is the user's 'userName' that has sent the message
// To is a slice containing usernames of whom the Sender want to send this message to
type ClientSendMessage struct {
	TimeStamp time.Time `json:"timeStamp"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
	To        []string  `json:"To"`
}

// ClientReceiveMessage is the json struct that server uses to send clients messages to clients
// clients should get message in this json format
// TimeStamp is the time that the Sender has sent the message
// Text is the Sender's text message
// Sender is the Sender's 'userName' that has sent the message
type ClientReceiveMessage struct {
	TimeStamp time.Time `json:"timeStamp"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
}
