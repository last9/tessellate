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

type Status int32

const (
	Status_INACTIVE Status = 0
	Status_ACTIVE   Status = 1
)

type Layout struct {
	Id     string
	Plan   map[string]interface{}
	Status Status
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

type JobState int32

const (
	JobState_PENDING JobState = 0
	JobState_RUNNING JobState = 1
	JobState_FAILED  JobState = 2
	JobState_ABORTED JobState = 3
	JobState_DONE    JobState = 4
	JobState_ERROR   JobState = 5
)

type Operation int32

const (
	Operation_APPLY   Operation = 0
	Operation_DESTROY Operation = 1
)

type Job struct {
	Id            string
	LayoutId      string
	LayoutVersion string
	Status        JobState
	VarsId        string
	VarsVersion   string
	Op            Operation
	Dry           bool
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
