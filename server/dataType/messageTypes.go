package dataType

import "time"

const (
	Received        = "RCV"
	Approve         = "APV"
	InvalidUserName = "IUN"
	NoSuchUser      = "NSU"
	AlreadyReg      = "ART"
	InvalidAuth     = "IAT"
)

type Response struct {
	Value string `json:"value"`
}

type Authentication struct {
	ClientID []byte `json:"clientID"`
	UserName string `json:"userName"`
}

type Registration struct {
	ClientID []byte `json:"clientID"`
	UserName string `json:"userName"`
	Name     string `json:"name"`
}

type ClientSendMessage struct {
	TimeStamp time.Time `json:"timeStamp"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
	To        []string  `json:"To"`
}

type ClientReceiveMessage struct {
	TimeStamp time.Time `json:"timeStamp"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
}
