package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
)

// everything starts from here
// database connection, controller, servers mux, mux handlers and finally server starts to listen
func main() {

	defer fmt.Println("Server stopped working!")

	db, err := createDBConnection("mahdi:123@/redFokDB?parseTime=true")
	if err != nil || db == nil {
		logError("createDBConnection", err)
		return
	}

	controller := initNewController(*db)
	mux := http.NewServeMux()
	gate := &processGate{isGateOpen: true}
	go func() { gate.dbConnWatcher(controller) }()

	mux.Handle("/messaging", websocket.Handler(
		func(conn *websocket.Conn) {
			if !gate.pGateCheck() {
				_ = conn.Close()
				return
			}

			controller.messenger(conn)
		}))

	mux.Handle("/registration", websocket.Handler(
		func(conn *websocket.Conn) {
			if !gate.pGateCheck() {
				_ = conn.Close()
				return
			}

			controller.register(conn)
		}))

	mux.Handle("/deletion", websocket.Handler(
		func(conn *websocket.Conn) {
			if !gate.pGateCheck() {
				_ = conn.Close()
				return
			}

			controller.deleter(conn)
		}))

	server := http.Server{
		Addr:    ":13013",
		Handler: mux,
	}

	fmt.Println("Server is running . . .")
	err = server.ListenAndServe()
	if err != nil {
		logError("ListenAndServe", err)
		return
	}
}
