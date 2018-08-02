package main

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/storage/types"
)

// Plan that is about to be saved will be appended with terraform layout.
func padLayoutWithProvider(plan map[string]json.RawMessage, vars *types.Vars) error {
	if vars == nil || len(*vars) == 0 {
		return nil
	}

	b, err := vars.Marshal()
	if err != nil {
		return errors.Wrap(err, "Cannot marshal providers from worksapce")
	}

	plan["provider_defaults"] = json.RawMessage(b)
	return nil
}
