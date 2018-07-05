package server

import "github.com/oogway/tessellate/storage"

type Server struct {
	store storage.Storer
}

func New(store storage.Storer) TessellateServer {
	return &Server{}
}
