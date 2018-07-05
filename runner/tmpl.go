package runner

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"
	"github.com/flosch/pongo2"
)

func tmplVars(m interface{}) (map[string]pongo2.Context, error) {
	if m == nil {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	x := map[string]pongo2.Context{}
	if err := json.Unmarshal(b, &x); err != nil {
		return nil, err
	}

	return x, nil
}

// Check if the bytes will yield a hash
func isJSON(b []byte) bool {
	var js map[string]interface{}
	err := json.Unmarshal(b, &js)
	return err == nil
}

// Parse the given bytes for the vars supplied.
// Return rendered bytes, only if they will yield a valid hash.
func parseLayout(data json.RawMessage, vars pongo2.Context) ([]byte, error) {
	var x string
	if err := json.Unmarshal(data, &x); err != nil {
		x = string(data)
	}

	tpl, err := pongo2.FromString(x)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create template from String")
	}

	//ctx := pongo2.Context{}
	//if vars != nil {
	//	for k, v := range vars {
	//		ctx[k] = v
	//	}
	//}

	out, err := tpl.Execute(vars)
	if err != nil {
		log.Println(out)
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
