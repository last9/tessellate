package server

import "gitlab.com/tsocial/sre/tessellate/storage"

type Server struct {
	store storage.Storer
}

func New(store storage.Storer) TessellateServer {
	return &Server{store: store}
}
