package dispatcher

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	cl := NewNomadClient(NomadConfig{
		Address:    "http://127.0.0.1:4646",
		Datacenter: "dc1",
		Image:      "redis",
		CPU:        "32",
		Memory:     "32",
	})

	Set(cl)
	os.Exit(m.Run())
}

func TestClient_Dispatch(t *testing.T) {
	assert.Nil(t, Get().Dispatch("job", "workspace"))
}
