package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"strings"
	"time"
)

type sendMessage struct {
	TimeStamp time.Time `json:"timeStamp"`
	Text      string    `json:"text"`
	To        []string  `json:"To"`
}

type receiveMessage struct {
	TimeStamp time.Time `json:"timeStamp"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
}

type authentication struct {
	ClientID []byte `json:"clientID"`
	UserName string `json:"userName"`
}

type response struct {
	Value string `json:"value"`
}

type registration struct {
	ClientID []byte `json:"clientID"`
	UserName string `json:"userName"`
	Name     string `json:"name"`
}

var (
	userName string
	users    []string
)

//This client is for testing process and should be developed for real use later

func main() {

	op := flag.String("op", "m", "m, r, d, t ...")
	flag.Parse()

	switch strings.ToLower(*op) {
	case "r":
		registering()
	case "m":
		messaging()
	case "d":
		deletion()
	case "t":
		test()
	}
}

func test() {

	fmt.Println("Test")
}

func deletion() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter your clientID: ")
	scanner.Scan()
	clientID := scanner.Text()
	fmt.Print("Enter your username: ")
	scanner.Scan()
	userName = scanner.Text()

	conn, err := websocket.Dial("ws://localhost:13013/api/deletion", "", "http://test")
	if err != nil {
		log.Fatalln(err)
	}

	hash := sha1.New()
	hash.Write([]byte(clientID))
	hashedClientID := hash.Sum(nil)
	err = websocket.JSON.Send(conn, authentication{
		ClientID: hashedClientID,
		UserName: userName,
	})
	if err != nil {
		log.Fatalln(err)
	}

	var res response
	err = websocket.JSON.Receive(conn, &res)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(res)
}

func registering() {

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter your clientID: ")
	scanner.Scan()
	clientID := scanner.Text()
	fmt.Print("Enter your username: ")
	scanner.Scan()
	userName = scanner.Text()
	fmt.Print("Enter your name: ")
	scanner.Scan()
	name := scanner.Text()

	conn, err := websocket.Dial("ws://localhost:13013/api/registration", "", "http://test")
	if err != nil {
		log.Fatalln(err)
	}

	hash := sha1.New()
	hash.Write([]byte(clientID))
	hashedClientID := hash.Sum(nil)

	err = websocket.JSON.Send(conn, registration{
		ClientID: hashedClientID,
		UserName: userName,
		Name:     name,
	})
	if err != nil {
		log.Fatalln(err)
	}

	var res response
	err = websocket.JSON.Receive(conn, &res)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(res)
}

func messaging() {

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter your clientID: ")
	scanner.Scan()
	clientID := scanner.Text()
	fmt.Print("Enter your username: ")
	scanner.Scan()
	userName = scanner.Text()
	fmt.Print("Enter the users to send message: ")
	var allUsers string
	_, _ = fmt.Scan(&allUsers)
	users = strings.Split(allUsers, "-")

	conn, err := websocket.Dial("ws://localhost:13013/api/messaging", "", "http://test")
	if err != nil {
		log.Fatalln(err)
	}

	hash := sha1.New()
	hash.Write([]byte(clientID))
	hashedClientID := hash.Sum(nil)
	err = websocket.JSON.Send(conn, authentication{
		ClientID: hashedClientID,
		UserName: userName,
	})
	if err != nil {
		log.Fatalln(err)
	}

	var res response
	err = websocket.JSON.Receive(conn, &res)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(res)

	if res.Value == "APV" {
		var sendNum, receiveNum int

		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				if scanner.Text() == "exit" {
					fmt.Println("Number of send: ", sendNum)
					fmt.Println("Number of receive: ", receiveNum)
					os.Exit(0)
				}
			}
		}()

		go receiving(conn, &receiveNum)

		sending(conn, &sendNum)
	}
}

func receiving(conn *websocket.Conn, num *int) {

	fmt.Println("Receiving started . . .")

	for {
		var data []byte
		err := websocket.Message.Receive(conn, &data)
		if err != nil {
			log.Fatalln(err)
		}

		var res response
		err = json.Unmarshal(data, &res)
		if err != nil {
			log.Fatalln(err)
		}
		if res.Value == "" {
			var rec receiveMessage
			err = json.Unmarshal(data, &rec)
			if err != nil {
				log.Fatalln(err)
			}
			rec.TimeStamp = rec.TimeStamp.In(time.Local)
			fmt.Println(rec)
			*num++
			continue
		}
		fmt.Println(res)
	}
}

func sending(conn *websocket.Conn, num *int) {

	fmt.Println("Sending started . . .")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()

		err := websocket.JSON.Send(conn, sendMessage{
			TimeStamp: time.Now(),
			Text:      text,
			To:        users,
		})
		if err != nil {
			fmt.Println("Error in Send Data, ", err)
		}

		*num++
	}
}
