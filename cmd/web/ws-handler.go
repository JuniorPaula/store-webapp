package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	*websocket.Conn
}

type WSPayload struct {
	Action      string               `json:"action"`
	Message     string               `json:"message"`
	MessageType string               `json:"message_type"`
	UserName    string               `json:"user_name"`
	Conn        *WebSocketConnection `json:"-"`
}

type WSJSONResponse struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var clients = make(map[WebSocketConnection]string)
var wsChan = make(chan WSPayload)

func (app *application) WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println("Client Connected to Endpoint:", r.RemoteAddr)

	var response WSJSONResponse
	response.Message = "Connected to server"

	err = ws.WriteJSON(response)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	conn := &WebSocketConnection{Conn: ws}
	clients[*conn] = ""

	go app.ListenForWS(conn)
}

func (app *application) ListenForWS(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			app.errorLog.Println("ERROR:", fmt.Sprintf("%v", r))
		}
	}()

	var payload WSPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		payload.Conn = conn

		wsChan <- payload
	}
}

func (app *application) ListenToWsChannel() {
	var response WSJSONResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "DELETE_USER":
			response.Action = "logout"
			response.Message = "User deleted, logging out!"
			app.broadcastToAll(response)

		default:
		}
	}
}

func (app *application) broadcastToAll(response WSJSONResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			app.errorLog.Println(err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
