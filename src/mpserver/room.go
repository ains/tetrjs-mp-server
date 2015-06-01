package main

import (
	"encoding/json"
	"math/rand"
	"time"
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

	startGame chan *player

	seed int32
}

func newRoom(id string, host *player) *room {
	room := &room{
		id:           id,
		addPlayer:    make(chan *player),
		removePlayer: make(chan *connection),
		startGame:    make(chan *player),
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
		case _ = <-r.startGame:
			r.seed = rand.Int31()

		}
	}
}

func (r *room) GetPlayerForConnection(c *connection) *player {
	return r.players[c]
}

func (r *room) BroadcastMessage(type_ string, data interface{}) {
	message := make(map[string]interface{})
	message["type"] = type_
	message["data"] = data

	jsonStr, _ := json.Marshal(message)

	for c := range r.players {
		c.send <- jsonStr
	}
}

func (r *room) Close() {

}
