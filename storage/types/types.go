package types

import (
	"path"

	"encoding/json"

	"github.com/satori/go.uuid"
)

type Tree struct {
	Name     string
	TreeType string
	child    *Tree
}

func (n *Tree) MakePath() string {
	d := path.Join(n.TreeType, n.Name)
	if n.child != nil {
		d = path.Join(d, n.child.MakePath())
	}
	return d
}

type ReaderWriter interface {
	MakePath(tree *Tree) string
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Workspace string

func (w *Workspace) MakePath(_ *Tree) string {
	return path.Join("workspace", string(*w))
}

func (w *Workspace) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

func (w *Workspace) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

type Vars map[string]interface{}

func (v *Vars) MakePath(n *Tree) string {
	return path.Join(n.MakePath(), "vars")
}

func (w *Vars) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

func (w *Vars) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

type Layout struct {
	Id     string
	Plan   map[string]interface{}
	Status string
}

func (l *Layout) MakePath(n *Tree) string {
	return path.Join(n.MakePath(), "layout", l.Id)
}

func (w *Layout) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

func (w *Layout) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

func MakeVersion() string {
	return uuid.NewV4().String()
}

/*
const (
	INACTIVE = iota // a == 1 (iota has been reset)
	ACTIVE   = iota // b == 2
)
*/

type Job struct {
	Id       string
	LayoutId string
	Status   string
	VarsId   string
	Op       string
}

func (v *Job) MakePath(n *Tree) string {
	return path.Join(n.MakePath(), "jobs", v.Id)
}

func (w *Job) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

func (w *Job) Marshal() ([]byte, error) {
	return json.Marshal(w)
}
