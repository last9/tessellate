package tmpl

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"
	"github.com/flosch/pongo2"
)

// Check if the bytes will yield a hash
func isJSON(b []byte) bool {
	var js interface{}
	err := json.Unmarshal(b, &js)
	return err == nil
}

// Parse the given bytes for the vars supplied.
// Return rendered bytes, only if they will yield a valid hash.
func ParseLayout(data json.RawMessage, vars pongo2.Context) ([]byte, error) {
	var x string
	if err := json.Unmarshal(data, &x); err != nil {
		x = string(data)
	}

	out, err := Parse(x, vars)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot parse template.")
	}

	outBytes := []byte(out)
	if !isJSON(outBytes) {
		log.Println(string(data))
		log.Println(vars)
		log.Println(out)

		return nil, errors.New("Rendered layout is not a valid JSON.")
	}

	return outBytes, err
}

func Parse(_tmpl string, data pongo2.Context) (string, error) {
	tpl, err := pongo2.FromString(_tmpl)
	if err != nil {
		return "", errors.Wrap(err, "Cannot create template from String")
	}

	return tpl.Execute(data)
}
