package main

import "github.com/ains/gotetris"

type player struct {
	board      *gotetris.Game
	connection *connection
}
