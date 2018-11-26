package commons

import (
	"testing"

	"os"

	"strings"

	"github.com/stretchr/testify/assert"
)

func TestCandidateFiles(t *testing.T) {
	t.Run("Check default list of files, no manifest", func(t *testing.T) {
		f, err := CandidateFiles("testdata", nil)
		assert.Nil(t, err)

		assert.Equal(t, f, []string{"testdata/hello.tf.json"})
	})

	t.Run("Check custom list of files, as per the manifest data", func(t *testing.T) {
		f, err := CandidateFiles("testdata", []string{".txt"})
		assert.Nil(t, err)

		assert.Equal(t, f, []string{"testdata/a.txt", "testdata/hello.tf.json"})
	})

	t.Run("Should return both .tmpl and tf.json files as per manifest", func(t *testing.T) {
		f, err := CandidateFiles("testdata", []string{".tmpl"})
		assert.Nil(t, err)
		assert.Equal(t, 2, len(f))
	})
}

func TestReadFileLines(t *testing.T) {
	t.Run("Should return manifest file contents and return extensions as an array of string.", func(t *testing.T) {
		expected := []string{".txt", ".tmpl"}
		lines, err := ReadFileLines("testdata/.tsl8")
		assert.Nil(t, err)
		assert.Equal(t, 2, len(lines))
		assert.Equal(t, strings.Join(lines, ","), strings.Join(expected, ","))
	})

	t.Run("Should read a single line with no new line character", func(t *testing.T) {
		expected := []string{".tmpl"}
		lines, err := ReadFileLines("testdata/newline.tsl8")
		assert.Nil(t, err)
		assert.Equal(t, len(expected), len(lines))
		assert.Equal(t, strings.Join(lines, ","), strings.Join(expected, ","))
	})
}

func TestFileType(t *testing.T) {
	t.Run("Should return actual file extensions and not txt", func(t *testing.T) {
		fileType, err1 := GetFileContentType("testdata/presentation.txt")
		assert.Nil(t, err1)
		assert.Equal(t, "zip", fileType)

		fileType2, err2 := GetFileContentType("testdata/movie.txt")
		assert.Nil(t, err2)
		assert.Equal(t, "mp4", fileType2)

		fileType3, err3 := GetFileContentType("testdata/main.txt")
		assert.Nil(t, err3)
		assert.Equal(t, "elf", fileType3)
	})

	t.Run("Should identify all text files irrespective of extensions.", func(t *testing.T) {
		fileType, err1 := GetFileContentType("testdata/a.txt")
		assert.Nil(t, err1)
		assert.Equal(t, "unknown", fileType)
		fileType2, err2 := GetFileContentType("testdata/newline.tsl8")
		assert.Nil(t, err2)
		assert.Equal(t, "unknown", fileType2)

		fileType3, err3 := GetFileContentType("testdata/hello.tf.json")
		assert.Nil(t, err3)
		assert.Equal(t, "unknown", fileType3)
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
