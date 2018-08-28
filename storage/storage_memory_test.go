// +build !integration

package storage

import (
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/tsocial/tessellate/storage/memory"
	"github.com/tsocial/tessellate/utils"
)

func TestMain(m *testing.M) {
	//Seed Random number generator.
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	bucket := utils.RandString(8)
	store = memory.MakeBoltStore(bucket, "/tmp/"+bucket)
	store.Setup()

	os.Exit(func() int {
		defer store.DeleteKeys(bucket)
		return m.Run()
	}())
}
