package runner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/flosch/pongo2"
	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/tmpl"
)

const (
	ApplyOp   = 0
	DestroyOp = 1
	PlanOp    = 2
)

var opMap = map[int32][]string{
	PlanOp:    []string{"plan"},
	ApplyOp:   []string{"apply", "-auto-approve"},
	DestroyOp: []string{"destroy", "-auto-approve"},
}

func remoteLayout(addr, path string) map[string]interface{} {
	return map[string]interface{}{
		"terraform": map[string]interface{}{
			"backend": map[string]interface{}{
				"consul": map[string]interface{}{
					"address": addr,
					"path":    path,
				},
			},
		},
	}
}

func tmplVars(m interface{}) (map[string]pongo2.Context, error) {
	if m == nil {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	x := map[string]pongo2.Context{}
	if err := json.Unmarshal(b, &x); err != nil {
		return nil, err
	}

	return x, nil
}

// Just a wrapper around Make Directory.
func makeDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

type Cmd struct {
	key        string
	skipInit   bool
	op         []string
	layout     map[string]json.RawMessage
	vars       map[string]interface{}
	stdout     OutWriteCloser
	stderr     OutWriteCloser
	dir        string
	logPrefix  string
	remoteAddr string
	remotePath string
}

// - Prepares the Basic Directories.
// - Sets the stdout and stderr file.
// - Creates the Command.
// - Finally, runs the command.
func (p *Cmd) Run() error {
	if err := makeDir(p.dir); err != nil {
		return errors.Wrapf(err, "Cannot make Directory %v", p.dir)
	}

	p.stdout = NewOutWriterCloser(false, p.logPrefix, "stdout")
	defer p.stdout.Close()

	p.stderr = NewOutWriterCloser(true, p.logPrefix, "stderr")
	defer p.stderr.Close()

	if err := p.saveLayout(); err != nil {
		return errors.Wrap(err, "Cannot save Layout")
	}

	if err := p.saveVars(); err != nil {
		return errors.Wrap(err, "Cannot save vars")
	}

	if p.remoteAddr != "" && p.remotePath != "" {
		if err := p.saveRemote(p.remoteAddr); err != nil {
			return errors.Wrap(err, "Cannot save Remote")
		}
	}

	if err := p.initLayout(); err != nil {
		return errors.Wrap(err, "Cannot init Layout")
	}

	c := p.getCmd()
	p.stdout.Output(fmt.Sprintf("Executing Command %+v", c))

	if err := c.Run(); err != nil {
		return errors.Wrap(err, "Error executing Command")
	}

	return nil
}

func (p *Cmd) SetRemotePath(path string) {
	p.remotePath = path
}

func (p *Cmd) GetStderr() OutWriteCloser {
	return p.stderr
}

// Command Setter
func (p *Cmd) SetOp(op int32) {
	o, ok := opMap[op]
	if !ok {
		o = opMap[ApplyOp]
	}

	p.op = o
}

// Base-Directory Setter.
func (p *Cmd) SetDir(name string) {
	p.dir = name
}

// Layout Setter.
func (p *Cmd) SetLayout(v map[string]json.RawMessage) {
	p.layout = v
}

func (p *Cmd) SetRemote(addr string) {
	p.remoteAddr = addr
}

// SetVars is vars Setter.
func (p *Cmd) SetVars(v map[string]interface{}) {
	p.vars = v
}

func (p *Cmd) SetLogPrefix(prefix string) {
	p.logPrefix = prefix
}

func (p *Cmd) ClearDir(path string) error {
	if p.dir != "" {
		return os.RemoveAll(path)
	}

	return nil
}

// Save the remote layout in a directory called .terraform.
func (p *Cmd) saveRemote(addr string) error {
	lPath := fmt.Sprintf("%v/state.tf.json", p.dir)
	p.stdout.Output("Saving Remote state file")
	lData, err := json.Marshal(remoteLayout(addr, p.remotePath))
	if err != nil {
		return errors.Wrap(err, "Cannot Marshal remote State")
	}

	if err := ioutil.WriteFile(lPath, lData, os.ModePerm); err != nil {
		p.stdout.Output("Cannot save tfstate")
		return errors.Wrap(err, "Cannot save tfstate to file")
	}

	return nil
}

// Layouts is a json of { key: layout_map }. Save them all to keyname.tf.json
func (p *Cmd) saveLayout() error {
	tv, err := tmplVars(p.vars["__tmpl__vars__"])
	if err != nil {
		return err
	}

	fn := func(data json.RawMessage, name string) error {
		p.stdout.Output(fmt.Sprintf("Saving Layout %v", name))
		layout, err := tmpl.ParseLayout(data, tv[name])
		if err != nil {
			return errors.Wrap(err, "Invalid Layout template")
		}

		lPath := fmt.Sprintf("%v/%v.tf.json", p.dir, name)
		if err := ioutil.WriteFile(lPath, layout, os.ModePerm); err != nil {
			p.stdout.Output("Cannot save layout file.")
			return errors.Wrap(err, "Cannot save Layout file")
		}

		return nil
	}

	for name, content := range p.layout {
		if err := fn(content, name); err != nil {
			return err
		}
	}

	return nil
}

// Save the variables file before starting the process.
func (p *Cmd) saveVars() error {
	// Do not save empty variables.
	if len(p.vars) == 0 {
		p.stdout.Output("Skipping empty Vars file")
		return nil
	}

	p.stdout.Output(fmt.Sprintf("%+v", p.vars))
	// Make a Copy of Variables before making alterations.
	vars, err := DeepCopy(p.vars)
	if err != nil {
		return errors.Wrap(err, "Cannot deep Copy vars")
	}

	delete(vars, "__tmpl__vars__")

	p.stdout.Output("Saving Vars file")
	lData, err := json.Marshal(vars)
	if err != nil {
		return errors.Wrap(err, "Cannot Marshal vars")
	}

	lPath := fmt.Sprintf("%v/terraform.tfvars.json", p.dir)
	if err := ioutil.WriteFile(lPath, lData, os.ModePerm); err != nil {
		p.stdout.Output("Cannot save vars file")
		return errors.Wrap(err, "Cannot save vars to file")
	}

	return nil
}

func (p *Cmd) getCmd() *exec.Cmd {
	op := append(p.op, "-no-color")
	cmd := exec.Command(TerraformPath(), op...)
	cmd.Stdout = p.stdout
	cmd.Stderr = p.stderr

	// Specify the working directory.
	cmd.Dir = p.dir

	return cmd
}

func (p *Cmd) initLayout() error {
	c := exec.Command(TerraformPath(), "init", "-input=false", "-no-color")

	//if config.Base.TerraformPluginDir() != "" {
	//	c.Args = append(c.Args, fmt.Sprintf("-plugin-dir=%v", config.Base.TerraformPluginDir()))
	//}

	// Specify the working directory.
	c.Stderr = p.stderr
	c.Dir = p.dir

	log.Printf("Executing %+v", c)
	_, err := c.Output()
	if err != nil {
		log.Println(p.stderr.GetBuffer())
		return errors.Wrap(err, "Error executing init")
	}

	return nil
}

func TerraformPath() string {
	return "terraform"
}
