package main

import (
	"encoding/json"

	"github.com/ains/gotetris"
)
import "code.google.com/p/go-uuid/uuid"

type player struct {
	id           string
	board        gotetris.Game
	pieceBag     *gotetris.PieceBag
	currentPiece int
	connection   *connection
}

func newPlayer(connection *connection) *player {
	return &player{
		id:         uuid.NewRandom().String(),
		board:      gotetris.Game{},
		connection: connection,
	}
}

func (p *player) SendMessage(messageType string, data interface{}) {
	message := make(map[string]interface{})
	message["type"] = messageType
	message["data"] = data

	jsonStr, _ := json.Marshal(message)
	p.connection.send <- jsonStr
}
