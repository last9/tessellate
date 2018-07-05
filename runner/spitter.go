package runner

import (
	"io"
	"log"
	"strings"
)

type OutWriteCloser interface {
	io.WriteCloser
	Output(string)
	GetBuffer() []string
}

func makeOutWriteCloser(prefix string) *spitter {
	x := spitter{prefix: prefix}
	return &x
}

type spitter struct {
	prefix   string
	buffered bool
	lines    []string
}

func (s *spitter) GetBuffer() []string {
	return s.lines
}

func (s *spitter) Write(l []byte) (int, error) {
	out := string(l)
	s.Output(out)
	return len(l), nil
}

func (s *spitter) Output(out string) {
	if strings.HasSuffix(out, "\n") {
		out = strings.TrimRight(out, "\n")
	}

	if s.buffered {
		s.lines = append(s.lines, out)
	}

	log.Printf("[%v] %v", s.prefix, out)
}

func (s *spitter) Close() error {
	return nil
}

// Set the Output to a flat-file. Needs to be writeable to DB later.
func NewOutWriterCloser(buffered bool, idents ...string) OutWriteCloser {
	x := spitter{prefix: strings.Join(idents, " "), buffered: buffered}
	return &x
}
