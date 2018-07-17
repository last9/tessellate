package types

import (
	"path"

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
}

type Workspace string

func (w Workspace) Path(_ *Tree) string {
	return path.Join("workspace", string(w))
}

type Vars map[string]interface{}

func (v *Vars) Path(n *Tree) string {
	return path.Join(n.MakePath(), "vars")
}

type Layout struct {
	Id     string
	Plan   map[string]interface{}
	Status string
}

func (l *Layout) Path(n *Tree) string {
	return path.Join(n.MakePath(), "layout", l.Id)
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
