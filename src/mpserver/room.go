package main

import (
	"math/rand"
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
		message:      make(chan *roomMessage),
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
			player.SendMessage("roomDigest", r.GenerateRoomDigest())
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
		for _, p := range r.players {
			p.pieceBag = gotetris.NewPieceBag(int(r.seed), 7, &RNG{seed: int(r.seed)})
		}

		r.BroadcastMessage("gameStarted", map[string]interface{}{
			"seed": r.seed,
		})
	case m.message.MessageType == "move":
		pos := int(m.message.Data["position"].(float64))
		rot := int(m.message.Data["rotation"].(float64))

		player.board = gotetris.DropPiece(player.board,
			player.pieceBag.AtIndex(player.currentPiece), pos, rot)
		player.currentPiece++
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
	for _, p := range r.players {
		p.SendMessage(messageType, data)
	}
}

func (r *room) GenerateRoomDigest() map[string]interface{} {
	playerList := make([]interface{}, 0, len(r.players))
	for _, player := range r.players {
		playerList = append(playerList, map[string]string{
			"playerID": player.id,
		})
	}

	return map[string]interface{}{
		"players": playerList,
	}
}

func (r *room) Close() {

}
