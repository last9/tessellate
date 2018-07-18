package types

import (
	"path"

	"encoding/json"

	"github.com/satori/go.uuid"
)

const (
	WORKSPACE = "workspace"
	LAYOUT    = "layout"
)

// MakeTree populates a Tree based on Input.
// Tree in itself is quite vague (read generic), but consumption is specific to workspace
// and layouts.
// Example: MakeTRee(workspace_id) returns a tree for a Workspace
// whereas MakeTree(worksapce_id, layout_id) returns a tree for a Workspace and a Layout{.
func MakeTree(nodes ...string) *Tree {
	if len(nodes) < 1 {
		return &Tree{Name: "unknown", TreeType: "unknown"}
	}

	workspace := nodes[0]
	t := Tree{Name: workspace, TreeType: WORKSPACE}
	if len(nodes) > 1 {
		t.Child = &Tree{Name: nodes[1], TreeType: LAYOUT}
	}

	return &t
}

// Tree is a Hierarchial representation of a Path at which a node is expcted to be found.
type Tree struct {
	Name     string
	TreeType string
	Child    *Tree
}

func (n *Tree) MakePath() string {
	d := path.Join(n.TreeType, n.Name)
	if n.Child != nil {
		d = path.Join(d, n.Child.MakePath())
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
	Id     string                     `json:"id"`
	Plan   map[string]json.RawMessage `json:"plan"`
	Status string                     `json:"status"`
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

type Job struct {
	Id            string `json:"id"`
	LayoutId      string `json:"layout_id"`
	LayoutVersion string `json:"layout_version"`
	Status        string `json:"status"`
	VarsVersion   string `json:"vars_version"`
	Op            string `json:"op"`
	Dry           bool   `json:"dry"`
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
