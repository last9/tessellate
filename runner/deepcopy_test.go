package runner

import (
	"testing"
)

func TestDeepCopy(t *testing.T) {
	testCases := []struct {
		// original and expectedOriginal are the same value in each test case. We do
		// this to avoid unintentionally asserting against a mutated
		// expectedOriginal and having the test pass erroneously. We also do not
		// want to rely on the deep copy function we are testing to ensure this does
		// not happen.
		original         map[string]interface{}
		transformer      func(m map[string]interface{}) map[string]interface{}
		expectedCopy     map[string]interface{}
		expectedOriginal map[string]interface{}
	}{
		// reassignment of entire map, should be okay even without deepcopy.
		{
			original: nil,
			transformer: func(m map[string]interface{}) map[string]interface{} {
				return map[string]interface{}{}
			},
			expectedCopy:     map[string]interface{}{},
			expectedOriginal: nil,
		},
		{
			original: map[string]interface{}{},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				return nil
			},
			expectedCopy:     nil,
			expectedOriginal: map[string]interface{}{},
		},
		// mutation of map
		{
			original: map[string]interface{}{},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["foo"] = "bar"
				return m
			},
			expectedCopy: map[string]interface{}{
				"foo": "bar",
			},
			expectedOriginal: map[string]interface{}{},
		},
		{
			original: map[string]interface{}{
				"foo": "bar",
			},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["foo"] = "car"
				return m
			},
			expectedCopy: map[string]interface{}{
				"foo": "car",
			},
			expectedOriginal: map[string]interface{}{
				"foo": "bar",
			},
		},
		// mutation of nested maps
		{
			original: map[string]interface{}{},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["foo"] = map[string]interface{}{
					"biz": "baz",
				}
				return m
			},
			expectedCopy: map[string]interface{}{
				"foo": map[string]interface{}{
					"biz": "baz",
				},
			},
			expectedOriginal: map[string]interface{}{},
		},
		{
			original: map[string]interface{}{
				"foo": map[string]interface{}{
					"biz": "booz",
					"gaz": "gooz",
				},
			},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["foo"] = map[string]interface{}{
					"biz": "baz",
				}
				return m
			},
			expectedCopy: map[string]interface{}{
				"foo": map[string]interface{}{
					"biz": "baz",
				},
			},
			expectedOriginal: map[string]interface{}{
				"foo": map[string]interface{}{
					"biz": "booz",
					"gaz": "gooz",
				},
			},
		},
		// mutation of slice values
		{
			original: map[string]interface{}{
				"foo": []string{"biz", "baz"},
			},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["foo"].([]string)[0] = "hiz"
				return m
			},
			expectedCopy: map[string]interface{}{
				"foo": []string{"hiz", "baz"},
			},
			expectedOriginal: map[string]interface{}{
				"foo": []string{"biz", "baz"},
			},
		},
		{
			original: map[string]interface{}{
				"foo": []string{"biz", "baz"},
				"foo2": map[string]interface{}{
					"foo3": []string{"123"},
				},
			},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["foo"].([]string)[0] = "hiz"
				return m
			},
			expectedCopy: map[string]interface{}{
				"foo": []string{"hiz", "baz"},
				"foo2": map[string]interface{}{
					"foo3": []string{"123"},
				},
			},
			expectedOriginal: map[string]interface{}{
				"foo": []string{"biz", "baz"},
				"foo2": map[string]interface{}{
					"foo3": []string{"123"},
				},
			},
		},
	}

	for _, tc := range testCases {
		_, err := DeepCopy(tc.original)
		if err != nil {
			t.Error(err)
			return
		}
	}
}
