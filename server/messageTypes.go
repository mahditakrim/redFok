package main

import "time"

// these are the flags that we use to send server responses.
// the values contains three digits of the flag name.
// received is the flag that server uses to response to clients saying that we processed your message successfully.
// approved is the flag that server uses to response to clients saying that your operation processed successfully.
// invalidUserName is the flag that server uses to response to clients saying that username is not valid to be save in database.
// noSuchUser is the flag that server uses to response to clients saying that the username you want to send message to it, is not existing in database.
// alreadyReg is the flag that server uses to response to clients saying that the ClientID is already existing in database.
// invalidAuth is the flag that server uses to response to clients saying that authentication is not valid.
const (
	received        = "RCV"
	approved        = "APV"
	invalidUserName = "IUN"
	noSuchUser      = "NSU"
	alreadyReg      = "ART"
	invalidAuth     = "IAT"
)

// response is the json struct that we use to send server responses.
// Value can be either one of the above flags.
type response struct {
	Value string `json:"value"`
}

// authentication is the json struct that clients should send at the very beginning of connection.
// ClientID is the unique identifier that we use for 2FA and ... .
// UserName is the unique identifier that we use to detect different users from each other.
type authentication struct {
	ClientID []byte `json:"ClientID"`
	UserName string `json:"userName"`
}

// registration is the json struct that clients should use for sending info to register an account.
// ClientID is the unique identifier that we use for 2FA and ... .
// UserName is the unique identifier that we use to detect different users from each other.
// Name is the optional name that user can choose for profile.
type registration struct {
	ClientID []byte `json:"ClientID"`
	UserName string `json:"userName"`
	Name     string `json:"name"`
}

// clientSendMessage is the json struct that clients should use for sending their messages.
// server processes client messages in this json format.
// TimeStamp is the time that user has sent the message.
// Text is user's text message.
// To is a slice containing usernames of whom the sender want to send this message to.
type clientSendMessage struct {
	TimeStamp time.Time `json:"timeStamp"`
	Text      string    `json:"text"`
	To        []string  `json:"To"`
}

// clientReceiveMessage is the json struct that server uses to send clients messages to clients.
// clients should get message in this json format.
// TimeStamp is the time that the sender has sent the message.
// Text is the sender's text message.
// Sender is the sender's 'userName' that has sent the message.
type clientReceiveMessage struct {
	TimeStamp time.Time `json:"timeStamp"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
}
