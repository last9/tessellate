package server

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/storage/types"
)

func TestValidationProvider(t *testing.T) {
	t.Run("Should identify a conflicting array provider", func(t *testing.T) {
		lBytes, err := ioutil.ReadFile("./testdata/file.tf.json")
		if err != nil {
			t.Error(err)
		}

		plan := map[string]json.RawMessage{}
		plan["provider_file.tf.json"] = uglyJson(lBytes)

		t.Run("Should return nil", func(t *testing.T) {
			assert.Nil(t, providerConflict(plan, nil))
		})

		t.Run("Should return nil for empty array", func(t *testing.T) {
			assert.Nil(t, providerConflict(plan, &types.Vars{}))
		})

		t.Run("Should not find a conflict for mismatch provider", func(t *testing.T) {
			v := &types.Vars{"alibaba": nil}
			assert.Nil(t, providerConflict(plan, v))
		})

		t.Run("Should complain about an overriding provider", func(t *testing.T) {
			v := &types.Vars{"aws": nil}
			assert.NotNil(t, providerConflict(plan, v))
		})
	})

	t.Run("Should identify a conflicting single provider", func(t *testing.T) {
		lBytes, err := ioutil.ReadFile("./testdata/file_single_provider.tf.json")
		if err != nil {
			t.Error(err)
		}

		plan := map[string]json.RawMessage{}
		plan["provider_file.tf.json"] = uglyJson(lBytes)

		t.Run("Should return nil", func(t *testing.T) {
			assert.Nil(t, providerConflict(plan, nil))
		})

		t.Run("Should return nil for empty array", func(t *testing.T) {
			assert.Nil(t, providerConflict(plan, &types.Vars{}))
		})

		t.Run("Should not find a conflict for mismatch provider", func(t *testing.T) {
			v := &types.Vars{"alibaba": nil}
			assert.Nil(t, providerConflict(plan, v))
		})

		t.Run("Should complain about an overriding provider", func(t *testing.T) {
			v := &types.Vars{"aws": nil}
			assert.NotNil(t, providerConflict(plan, v))
		})
	})
}
