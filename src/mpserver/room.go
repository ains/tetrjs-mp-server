package main

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/ains/gotetris"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type room struct {
	id      string
	players map[*connection]*player
	host    *player

	// New room creation message
	addPlayer chan *player

	// Player join room message
	removePlayer chan *connection

	message chan *roomMessage

	seed int32
}

type roomMessage struct {
	connection *connection
	message    *gameMessage
}

func newRoom(id string, host *player) *room {
	room := &room{
		id:           id,
		addPlayer:    make(chan *player),
		removePlayer: make(chan *connection),
		players:      make(map[*connection]*player),
		host:         host,
	}

	room.players[host.connection] = host

	return room
}

func (r *room) run() {
	for {
		select {
		case player := <-r.addPlayer:
			r.players[player.connection] = player
			r.BroadcastMessage("playerJoin", map[string]string{
				"playerID": player.id,
			})
		case c := <-r.removePlayer:
			player, ok := r.players[c]
			if ok {
				r.BroadcastMessage("playerLeave", map[string]string{
					"playerID": player.id,
				})
				delete(r.players, c)
			}
		case m := <-r.message:
			r.handleRoomMessage(m)
		}
	}
}

func (r *room) handleRoomMessage(m *roomMessage) {
	player := r.GetPlayerForConnection(m.connection)
	switch {
	case m.message.MessageType == "requestGameStart":
		r.seed = rand.Int31()
		r.BroadcastMessage("gameStarted", map[string]interface{}{
			"seed": r.seed,
		})
	case m.message.MessageType == "move":
		position := m.message.Data["position"]
		rotation := m.message.Data["rotation"]

		pos, _ := strconv.Atoi(position)
		rot, _ := strconv.Atoi(rotation)

		player.board = gotetris.DropPiece(player.board, gotetris.PieceMap['L'], pos, rot)
		player.board.OutputBoard()
		r.BroadcastMessage("playerMoved", map[string]interface{}{
			"playerID": player.id,
			"rot":      rot,
			"pos":      pos,
		})
	}
}

func (r *room) GetPlayerForConnection(c *connection) *player {
	return r.players[c]
}

func (r *room) BroadcastMessage(messageType string, data interface{}) {
	message := make(map[string]interface{})
	message["type"] = messageType
	message["data"] = data

	jsonStr, _ := json.Marshal(message)

	for c := range r.players {
		c.send <- jsonStr
	}
}

func (r *room) Close() {

}
