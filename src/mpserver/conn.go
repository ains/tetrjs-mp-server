package main

import (
	"net/http"

	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Current room this connection is in (can be nil)
	room *room

	hub *hub
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		messageParts := strings.Split(string(message), " ")
		if len(messageParts) > 0 {
			cmd := messageParts[0]
			fmt.Println(cmd)
			switch {
			// Joining, leaving and creating rooms are managed by the hub
			case cmd == "new":
				roomChan := make(chan *room)
				c.hub.newRoom <- &createRoomMsg{c, roomChan}
				c.room = <-roomChan
				if c.room != nil {
					c.send <- []byte(fmt.Sprintf("joined %s", c.room.id))
				}
			case cmd == "join":
				roomID := messageParts[1]
				roomChan := make(chan *room)
				c.hub.joinRoom <- &joinRoomMsg{c, roomID, roomChan}
				c.room = <-roomChan
				if c.room != nil {
					c.send <- []byte(fmt.Sprintf("joined %s", c.room.id))
				}
			case cmd == "leave":
				oldRoom := c.room
				if oldRoom != nil {
					successChan := make(chan bool)
					c.hub.leaveRoom <- &leaveRoomMsg{c, successChan}
					<-successChan
					c.room = nil
					c.send <- []byte(fmt.Sprintf("left %s", oldRoom.id))
				}
			default:
				// All other messages are handled by the room the player is in
				if c.room != nil {

				}
			}
		}
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type wsHandler struct {
	hub *hub
}

func (wsh wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, hub: wsh.hub}

	go c.writer()
	c.reader()
}
