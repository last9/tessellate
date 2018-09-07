package commons

import (
	"testing"

	"os"

	"strings"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestCandidateFiles(t *testing.T) {
	t.Run("Check default list of files", func(t *testing.T) {
		f, err := CandidateFiles("testdata", nil)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, f, []string{"testdata/hello.tf.json"})
	})

	t.Run("Check custom list of files", func(t *testing.T) {
		f, err := CandidateFiles("testdata", []string{".txt"})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, f, []string{"testdata/a.txt"})
	})

	t.Run("Should give both tmpl and tf.json files as per manifest", func(t *testing.T) {
		f, err := CandidateFiles("testdata", []string{".tmpl"})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(f)
	})
}

func TestReadFileLines(t *testing.T) {
	t.Run("Should return manifest file contents and return extensions an array of string.", func(t *testing.T) {
		expected := []string{".txt\n", ".tmpl"}
		lines, err := ReadFileLines("testdata/.tsl8")
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
		expected := []string{".tmpl\n"}
		lines, err := ReadFileLines("testdata/newline.tsl8")
		if err != nil {
			t.Fatal(err)
		}

		if len(expected) != len(lines) {
			t.Fatal("Expected 1 item in array, got %v", len(lines))
		}

		if strings.Join(lines, ",") != strings.Join(expected, ",") {
			t.Fatal("Expected %v, Read %v", expected, lines)
		}
	})
}

func TestFileType(t *testing.T) {
	t.Run("Should return actual file extensions and not txt", func(t *testing.T) {
		fileType, err := GetFileContentType("testdata/presentation.txt")
		if err != nil {
			t.Fatal(err)
		}
		if fileType != "zip" {
			t.Fatal("Expected to be a zip file, got %v", fileType)
		}

		fileType2, err := GetFileContentType("testdata/movie.txt")
		if err != nil {
			t.Fatal(err)
		}
		if fileType2 != "mp4" {
			t.Fatal("Expected to be a mp4 file, got %v", fileType2)
		}

		fileType3, err := GetFileContentType("testdata/main.txt")
		if err != nil {
			t.Fatal(err)
		}
		if fileType3 != "elf" {
			t.Fatal("Expected to be a mp4 file, got %v", fileType3)
		}
	})

	t.Run("Should identify all text files irrespective of extensions.", func(t *testing.T) {
		fileType, err := GetFileContentType("testdata/a.txt")
		if err != nil {
			t.Fatal(err)
		}
		if fileType != "unknown" {
			t.Fatal("Expected to be a text file, got %v", fileType)
		}

		fileType2, err := GetFileContentType("testdata/newline.tsl8")
		if err != nil {
			t.Fatal(err)
		}
		if fileType2 != "unknown" {
			t.Fatal("Expected to be a text file, got %v", fileType2)
		}

		fileType3, err := GetFileContentType("testdata/hello.tf.json")
		if err != nil {
			t.Fatal(err)
		}
		if fileType3 != "unknown" {
			t.Fatal("Expected to be a text file, got %v", fileType3)
		}

	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
