package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

type gameMessage struct {
	MessageType string            `json:"type"`
	Data        map[string]string `json:"data"`
}

func decodeMessage(data []byte) (*gameMessage, error) {
	message := &gameMessage{}
	err := json.Unmarshal(data, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (c *connection) reader() {
	for {
		_, messageStr, err := c.ws.ReadMessage()
		log.Println(string(messageStr))
		if err != nil {
			log.Println(err)
			break
		}

		message, err := decodeMessage(messageStr)
		if err != nil {
			log.Println(err)
			continue
		}

		messageType := message.MessageType
		fmt.Println(messageType)

		switch {
		// Joining, leaving and creating rooms are managed by the hub
		case messageType == "createRoom":
			roomChan := make(chan *room)
			c.hub.newRoom <- &createRoomMsg{c, roomChan}
			room := <-roomChan
			c.room = room
			room.BroadcastMessage("roomCreated", map[string]string{
				"roomID":   room.id,
				"playerID": room.host.id,
			})
		case messageType == "joinRoom":
			roomID, ok := message.Data["roomID"]
			if !ok {
				continue
			}
			roomChan := make(chan *room)
			c.hub.joinRoom <- &joinRoomMsg{c, roomID, roomChan}
			c.room = <-roomChan
		case messageType == "leaveRoom":
			oldRoom := c.room
			if oldRoom != nil {
				successChan := make(chan bool)
				c.hub.leaveRoom <- &leaveRoomMsg{c, successChan}
				<-successChan
				c.room = nil
			}
		default:
			// All other messages are handled by the room the player is in
			if c.room != nil {
				c.room.message <- &roomMessage{c, message}
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

var upgrader = &websocket.Upgrader{
	ReadBufferSize: 1024, WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsHandler struct {
	hub *hub
}

func (wsh wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, hub: wsh.hub}
	c.hub.register <- c
	defer func() {
		if c.room != nil {
			c.room.removePlayer <- c
		}
		c.hub.unregister <- c
	}()

	go c.writer()
	c.reader()
}
