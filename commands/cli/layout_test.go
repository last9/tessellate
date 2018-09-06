package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCandidateFiles(t *testing.T) {
	t.Run("Check default list of files", func(t *testing.T) {
		f, err := candidateFiles("testdata", nil)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, f, []string{"testdata/hello.tf.json"})
	})

	t.Run("Check custom list of files", func(t *testing.T) {
		f, err := candidateFiles("testdata", []string{".txt"})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, f, []string{"testdata/a.txt"})
	})
}
