package tmpl

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"encoding/json"
	"strings"
	"github.com/flosch/pongo2"
)

func TestTmpl(t *testing.T) {
	t.Run("Should render it correctly", func(t *testing.T) {
		d, err := ioutil.ReadFile("./testdata/layout.tmpl")
		assert.Nil(t, err)

		vars, err := ioutil.ReadFile("./testdata/vars.json")
		assert.Nil(t, err)

		y := pongo2.Context{}
		err2 := json.Unmarshal(vars, &y)
		assert.Nil(t, err2)

		out, err := parseLayout(d, y)
		assert.Nil(t, err)

		assert.Equal(t, true, strings.Contains(string(out), "012"))
	})

	t.Run("Should not render template", func(t *testing.T) {
		d, err := ioutil.ReadFile("./testdata/layout.tmpl")
		assert.Nil(t, err)

		y := pongo2.Context{}

		out, err := parseLayout(d, y)
		assert.Nil(t, err)

		assert.Equal(t, false, strings.Contains(string(out), "012"))
	})
}
