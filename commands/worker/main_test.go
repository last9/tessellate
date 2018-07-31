package main

import (
	"os"
	"testing"

	"github.com/tsocial/tessellate/storage/consul"
)

func TestMainRunner(t *testing.T) {
	// jID := "j123"
	// wID := "w123"
	// lID := "l123"

	store := consul.MakeConsulStore(os.Getenv("CONSUL_ADDR"))
	store.Setup()

	t.Run("Should get a valid cmd", func(t *testing.T) {
	})
}
