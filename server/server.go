package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
)

// everything starts from here.
// database connection, controller, servers mux, mux handlers and finally server starts to listen.
func main() {

	defer fmt.Println("Server stopped working!")

	dbConn, err := createDBConnection("mahdi:123@/redFokDB?parseTime=true")
	if err != nil || dbConn == nil {
		logError("createDBConnection", err)
		return
	}
	defer func() { _ = dbConn.db.Close() }()

	controller := initNewController(*dbConn)
	mux := http.NewServeMux()
	gate := &processGate{isGateOpen: true}
	go func() { gate.dbConnWatcher(controller) }()

	mux.Handle("/api/", websocket.Handler(
		func(conn *websocket.Conn) {

			if !gate.pGateCheck() {
				_ = conn.Close()
				return
			}

			switch conn.Request().RequestURI {
			case "/api/messaging":
				controller.messenger(conn)
			case "/api/registration":
				controller.register(conn)
			case "/api/deletion":
				controller.deleter(conn)
			}
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
