package main

type room struct {
	id      string
	players map[*connection]*player
	host    *player
}

func (r *room) AddPlayer(player *player) {
	r.players[player.connection] = player
}

func (r *room) RemovePlayer(c *connection) {
	delete(r.players, c)
}

func (r *room) GetPlayerForConnection(c *connection) *player {
	return r.players[c]
}

func (r *room) Close() {

}
