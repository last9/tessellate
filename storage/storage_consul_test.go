// +build integration

package storage

import (
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/tsocial/tessellate/storage/consul"
)

func TestMain(m *testing.M) {
	//Seed Random number generator.
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	store = consul.MakeConsulStore(os.Getenv("CONSUL_ADDR"))
	store.Setup()

	os.Exit(m.Run())
}
