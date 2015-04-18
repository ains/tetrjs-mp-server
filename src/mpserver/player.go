package main

import "github.com/ains/gotetris"

type player struct {
	board      *gotetris.Game
	connection *connection
}

func newPlayer(connection *connection) *player {
	return &player{
		board:      &gotetris.Game{},
		connection: connection,
	}
}
