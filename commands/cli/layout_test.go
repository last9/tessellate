package main

import (
	"testing"

	"os"

	"strings"

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

func TestReadFileLines(t *testing.T) {
	t.Run("Should return manifest file contents and return extensions an array of string.", func(t *testing.T) {
		expected := []string{".txt", ".tmpl"}
		lines, err := readFileLines("testdata/.tsl8")
		if err != nil {
			t.Fatal(err)
		}

		if len(lines) != 2 {
			t.Fatal("Expected 2 extensions to be read, found %v", len(lines))
		}

		if strings.Join(lines, ",") != strings.Join(expected, ",") {
			t.Fatal("Expected %v, Read %v", expected, lines)
		}
	})

	t.Run("Should read a single line with no new line character", func(t *testing.T) {
		expected := []string{".tmpl"}
		lines, err := readFileLines("testdata/newline.tsl8")
		if err != nil {
			t.Fatal(err)
		}
		if len(expected) != len(lines) {
			t.Fatal("Expected 1 item in array, got %v", len(lines))
		}
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
