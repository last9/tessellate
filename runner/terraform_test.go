package runner

import (
	"encoding/json"
	"log"
	"path"
	"testing"

	"io/ioutil"
)

var cmd *Cmd

func TestCmd_SetOp(t *testing.T) {
	cmd.SetOp(ApplyOp)
}

func TestCmd_SetDir(t *testing.T) {
	cmd.SetDir(path.Join("/tmp/test_runner"))
}

func TestCmd_SetVars(t *testing.T) {
}

func TestCmd_SetLayout(t *testing.T) {
	l := json.RawMessage{}
	d, err := ioutil.ReadFile("./testdata/sleep.tf.json")
	if err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(d, &l); err != nil {
		t.Fatal(err)
		return
	}

	out := map[string]json.RawMessage{}
	out["sleep"] = l
	cmd.SetLayout(out)
}

func TestCmd_ZRun(t *testing.T) {
	cmd.skipInit = true
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	cmd = &Cmd{}
	m.Run()
}
