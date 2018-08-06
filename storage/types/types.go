package types

import (
	"path"

	"encoding/json"

	"strings"

	"github.com/satori/go.uuid"
)

const (
	WORKSPACE = "workspaces"
	LAYOUT    = "layouts"
	JOB       = "jobs"
	VAR       = "vars"
	WATCH     = "watch"
	STATE     = "state"
)

var secretKeys = []string{"secret", "access"}

// MakeTree populates a Tree based on Input.
// Tree in itself is quite vague (read generic), but consumption is specific to workspace
// and layouts.
// Example: MakeTRee(workspace_id) returns a tree for a Workspace
// whereas MakeTree(workspace_id, layout_id) returns a tree for a Workspace and a Layout{.
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

type BaseType struct{}

func (b *BaseType) SaveId(string) {}

// Tree is a Hierarchical representation of a Path at which a node is expected to be found.
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
	SaveId(string)
}

type Workspace string

func (w *Workspace) SaveId(string) {}

func (w *Workspace) MakePath(_ *Tree) string {
	return path.Join(WORKSPACE, string(*w))
}

func (w *Workspace) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

func (w *Workspace) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

type Vars map[string]interface{}

func (v *Vars) SaveId(id string) {
	map[string]interface{}(*v)["id"] = id
}

func (v *Vars) MakePath(n *Tree) string {
	return path.Join(n.MakePath(), VAR)
}

func (w *Vars) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

func (w *Vars) RedactSecrets() {
	if len(*w) == 0 {
		return
	}

	mv := map[string]interface{}(*w)
	for k := range mv {
		mv[k] = redactSecrets(mv[k])
	}
}

func redactSecrets(m interface{}) interface{} {
	mp, ok := m.(map[string]interface{})
	if ok {
		for k := range mp {
			if _, ok := mp[k].(map[string]interface{}); ok {
				mp[k] = redactSecrets(mp[k])
				continue
			}

			if _, ok := mp[k].([]interface{}); ok {
				mp[k] = redactSecrets(mp[k])
				continue
			}

			_, ok = mp[k].(string)
			for _, s := range secretKeys {
				if strings.Contains(strings.ToLower(k), s) && ok {
					mp[k] = "***********"
				}
			}
		}
		return mp
	}

	ml, ok := m.([]interface{})
	if ok {
		for ix, i := range ml {
			if _, ok := i.(map[string]interface{}); ok {
				ml[ix] = redactSecrets(ml[ix])
			}
		}
		return ml
	}

	return m
}

func (w *Vars) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

type Layout struct {
	Id     string                     `json:"id"`
	Plan   map[string]json.RawMessage `json:"plan"`
	Status int32                      `json:"status"`
	*BaseType
}

func (l *Layout) SaveId(string) {}

func (l *Layout) MakePath(n *Tree) string {
	return path.Join(n.MakePath(), LAYOUT, l.Id)
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
	Status        int32  `json:"status"`
	VarsVersion   string `json:"vars_version"`
	Op            int32  `json:"op"`
	Dry           bool   `json:"dry"`
}

func (v *Job) SaveId(id string) {
	v.Id = id
}

func (v *Job) MakePath(n *Tree) string {
	return path.Join(n.MakePath(), JOB, v.LayoutId)
}

func (w *Job) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

func (w *Job) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

type Watch struct {
	Id         string
	SuccessURL string `json:"success_url"`
	FailureURL string `json:"failure_url"`
}

func (v *Watch) SaveId(id string) {
	v.Id = id
}

func (w *Watch) MakePath(n *Tree) string {
	return path.Join(n.MakePath(), WATCH)
}

func (w *Watch) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

func (w *Watch) Marshal() ([]byte, error) {
	return json.Marshal(w)
}
