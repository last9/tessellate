package runner

import (
	"fmt"
	"log"
	"testing"
)

type buffer struct {
	lines []string
}

func (b *buffer) Write(p []byte) (n int, err error) {
	b.lines = append(b.lines, string(p))
	return 1, nil
}

func TestSpitter(t *testing.T) {
	t.Run("Write should return length of line", func(t *testing.T) {
		s := spitter{prefix: "hello"}
		if l, e := s.Write([]byte("hello world")); e != nil {
			t.Error(e.Error())
			return
		} else if l != 11 {
			t.Errorf("Should have returned 11 length, Found %v", l)
			return
		}
	})

	t.Run("Buffered Logger should keep lines", func(t *testing.T) {
		s := &spitter{buffered: true}
		line1 := "hello 1"
		line2 := "hello 2"
		s.Output(line1)
		s.Output(line2)

		if s.lines[0] != line1 {
			t.Error("Output doesnt match")
		}

		if s.lines[1] != line2 {
			t.Error("Output doesnt match")
		}
	})

	t.Run("Output should return lines", func(t *testing.T) {
		b := &buffer{}
		log.SetOutput(b)

		s := &spitter{prefix: "prefix"}
		s.Output("hello world")
		fmt.Println(b.lines[0])
	})
}
