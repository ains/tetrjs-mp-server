package main

import "code.google.com/p/go-uuid/uuid"

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Created rooms
	rooms map[string]*room

	// New room creation message
	newRoom chan *createRoomMsg

	// New room creation message
	joinRoom chan *joinRoomMsg

	leaveRoom chan *leaveRoomMsg

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

type createRoomMsg struct {
	connection  *connection
	roomChannel chan *room
}

type joinRoomMsg struct {
	connection  *connection
	roomID      string
	roomChannel chan *room
}

type leaveRoomMsg struct {
	connection     *connection
	successChannel chan bool
}

func newHub() *hub {
	return &hub{
		newRoom:     make(chan *createRoomMsg),
		joinRoom:    make(chan *joinRoomMsg),
		leaveRoom:   make(chan *leaveRoomMsg),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
		rooms:       make(map[string]*room),
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.newRoom:
			newPlayer := newPlayer(m.connection)
			roomID := uuid.NewRandom().String()

			room := &room{id: roomID, host: newPlayer, players: make(map[*connection]*player)}
			room.AddPlayer(newPlayer)

			h.rooms[roomID] = room

			m.roomChannel <- room
		case m := <-h.joinRoom:
			player := newPlayer(m.connection)
			roomID := m.roomID
			room, exists := h.rooms[roomID]
			if exists {
				room.AddPlayer(player)
				m.roomChannel <- room
			} else {
				m.roomChannel <- nil
			}
		case m := <-h.leaveRoom:
			room := m.connection.room
			if room.host == room.GetPlayerForConnection(m.connection) {
				// Host has left, unregister the room
				delete(h.rooms, room.id)
				room.Close()
			} else {
				room.RemovePlayer(m.connection)
			}

			m.successChannel <- true
		}
	}
}
