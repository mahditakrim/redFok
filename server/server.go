package main

import (
	"fmt"
	"github.com/mahditakrim/redFok/server/database"
	"github.com/mahditakrim/redFok/server/handler"
	"github.com/mahditakrim/redFok/server/utility"
	"golang.org/x/net/websocket"
	"net/http"
)

func main() {

	defer fmt.Println("Server stopped working!")

	db, err := database.CreateDBConnection("mahdi:123@/redFokDB?parseTime=true")
	if err != nil || db == nil {
		utility.LogError("CreateDBConnection", err, true)
	}

	controller := handler.InitController(*db)
	mux := http.NewServeMux()
	gate := &processGate{isGateOpen: true}
	go func() { gate.dbConnWatcher(controller) }()

	mux.Handle("/messaging", websocket.Handler(
		func(conn *websocket.Conn) {
			if !gate.pGateCheck() {
				_ = conn.Close()
				return
			}

			controller.Messenger(conn)
		}))

	mux.Handle("/registration", websocket.Handler(
		func(conn *websocket.Conn) {
			if !gate.pGateCheck() {
				_ = conn.Close()
				return
			}

			controller.Register(conn)
		}))

	mux.Handle("/deletion", websocket.Handler(
		func(conn *websocket.Conn) {
			if !gate.pGateCheck() {
				_ = conn.Close()
				return
			}

			controller.Deleter(conn)
		}))

	server := http.Server{
		Addr:    ":13013",
		Handler: mux,
	}

	fmt.Println("Server is running . . .")
	err = server.ListenAndServe()
	if err != nil {
		utility.LogError("ListenAndServe", err, true)
	}
}
