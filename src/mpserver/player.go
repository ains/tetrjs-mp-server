package main

import "github.com/ains/gotetris"
import "code.google.com/p/go-uuid/uuid"

type player struct {
	id         string
	board      *gotetris.Game
	connection *connection
}

func newPlayer(connection *connection) *player {
	return &player{
		id:         uuid.NewRandom().String(),
		board:      &gotetris.Game{},
		connection: connection,
	}
}
