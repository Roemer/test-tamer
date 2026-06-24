package server

import "github.com/roemer/test-tamer/internal/store"

type Config struct {
	Address  string
	Version  string
	Subtitle string
	Store    store.Client
}
