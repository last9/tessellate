package dispatcher

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestClient_Dispatch(t *testing.T) {
	cl := NewNomadClient(NomadConfig{
		Address: "http://127.0.0.1:4646",
		Datacenter: "dc1",
		Image: "redis",
		CPU: "32",
		Memory: "32",
	})

	assert.Nil(t, cl.Dispatch("job", "workspace"))
}