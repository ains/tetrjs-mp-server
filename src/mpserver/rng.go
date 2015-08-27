package main

type RNG struct {
	seed int
}

func (rng *RNG) NextRandom() int {
	rng.seed = (rng.seed * 16807) % 2147483647
	return rng.seed
}
