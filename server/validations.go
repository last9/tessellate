package server

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"

	"path/filepath"

	"github.com/tsocial/tessellate/storage/types"
)

const (
	aliasKey    = "alias"
	providerKey = "provider"
)

// Loop over a Provider Map and look for any conflicting providers in the the Map.
func hasNonAliasProvider(in map[string]map[string]interface{}, p []string) string {
	for _, px := range p {
		pMap, ok := in[px]
		if !ok {
			continue
		}

		conflict := true
		for k := range pMap {
			if k != aliasKey {
				continue
			}

			conflict = false
			break
		}

		if conflict {
			return px
		}
	}

	return ""
}

// Walk the input variables recurisvely and search if there is a key called
// provider. If the key by same name is also found in worksapce Vars and has a corresponding
// non zero byte string, it implies that worksapce is shipping a default provider for that
// provider name and hence the layout cannot use a default provider. But at the same time,
// the provider supplied by the layout may have an alias value.
// Do not raise error if the provider in the layout input has an alias.
func providerConflict(input map[string]json.RawMessage, wvars *types.Vars) error {
	if wvars == nil {
		return nil
	}

	wProviders := []string{}
	for k := range *wvars {
		wProviders = append(wProviders, k)
	}

	if len(wProviders) == 0 {
		return nil
	}

	for fileName, layoutBytes := range input {
		if filepath.Ext(fileName) != ".json" {
			continue
		}
		layoutMap := map[string]json.RawMessage{}
		if err := json.Unmarshal(layoutBytes, &layoutMap); err != nil {
			return errors.Wrap(err, "Cannot parse files")
		}

		for k, v := range layoutMap {
			if k != providerKey {
				continue
			}

			amap := []map[string]map[string]interface{}{}
			if err := json.Unmarshal(v, &amap); err != nil {
				m := map[string]map[string]interface{}{}
				if err := json.Unmarshal(v, &m); err != nil {
					log.Println("Cannot unmarshal", string(v))
					continue
				}
				amap = append(amap, m)
			}

			for _, a := range amap {
				px := hasNonAliasProvider(a, wProviders)
				if px != "" {
					return errors.Errorf("%v is already provided by Workspace. Use an alias.", px)
				}
			}
		}
	}

	return nil
}
