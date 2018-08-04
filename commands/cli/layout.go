package main

import (
	"log"

	"os"

	"path/filepath"

	"fmt"

	"encoding/json"
	"io/ioutil"

	"strings"

	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/server"
	"gopkg.in/alecthomas/kingpin.v2"
)

type layout struct {
	id          string
	workspaceId string
	dirName     string
	dry         bool
	varsPath    string
}

func (cm *layout) layoutCreate(c *kingpin.ParseContext) error {
	if _, err := os.Stat(cm.dirName); err != nil {
		log.Printf("Directory '%s' does not exist\n", cm.dirName)
	}

	var files []string

	if err := filepath.Walk(cm.dirName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".json" || strings.Contains(path, "tfvars") {
			log.Printf("skipping %s", path)
			return nil
		}

		files = append(files, path)
		return nil

	}); err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no json file found in directory %s", cm.dirName)
	}

	fLayout := map[string]interface{}{}

	for _, f := range files {
		fBytes, err := ioutil.ReadFile(f)
		if err != nil {
			log.Println(err)
			return err
		}

		var fObj interface{}
		if err := json.Unmarshal(fBytes, &fObj); err != nil {
			log.Printf("invald json file: %s", f)
			return err
		}

		splits := strings.Split(f, "/")
		fLayout[splits[len(splits)-1]] = fObj
	}

	layoutBytes, err := json.Marshal(fLayout)
	if err != nil {
		log.Println(err)
		return err
	}

	req := &server.SaveLayoutRequest{
		Id:          cm.id,
		WorkspaceId: cm.workspaceId,
		Plan:        layoutBytes,
	}

	_, err = getClient().SaveLayout(makeContext(nil), req)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (cm *layout) layoutGet(c *kingpin.ParseContext) error {
	req := &server.LayoutRequest{
		Id:          cm.id,
		WorkspaceId: cm.workspaceId,
	}

	resp, err := getClient().GetLayout(makeContext(nil), req)
	if err != nil {
		return err
	}

	var plan interface{}
	if err := json.Unmarshal(resp.Plan, &plan); err != nil {
		return err
	}

	prettyPrint(plan)
	return nil
}

func (cm *layout) layoutApply(c *kingpin.ParseContext) error {
	vars, err := getVars(cm.varsPath)
	if err != nil {
		return err
	}

	req := &server.ApplyLayoutRequest{
		Id:          cm.id,
		WorkspaceId: cm.workspaceId,
		Vars:        vars,
		Dry:         cm.dry,
	}

	resp, err := getClient().ApplyLayout(makeContext(nil), req)
	if err != nil {
		return err
	}

	prettyPrint("JobID = " + resp.Id)
	return nil
}

func (cm *layout) layoutDestroy(c *kingpin.ParseContext) error {
	vars, err := getVars(cm.varsPath)
	if err != nil {
		return err
	}

	req := &server.DestroyLayoutRequest{
		Id:          cm.id,
		WorkspaceId: cm.workspaceId,
		Vars:        vars,
	}

	resp, err := getClient().DestroyLayout(makeContext(nil), req)
	if err != nil {
		return err
	}

	state := resp.Status

	prettyPrint(state)
	return nil
}

func (cm *layout) layoutStateGet(_ *kingpin.ParseContext) error {
	req := &server.GetStateRequest{
		LayoutId:    cm.id,
		WorkspaceId: cm.workspaceId,
	}

	resp, err := getClient().GetState(makeContext(nil), req)
	if err != nil {
		return err
	}

	var d interface{}
	if err := json.Unmarshal(resp.State, &d); err != nil {
		return err
	}

	prettyPrint(d)
	return nil
}

func getVars(path string) ([]byte, error) {
	if path == "" {
		return []byte{}, nil
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot read vars file.")
	}

	return b, nil
}

func addLayoutCommands(app *kingpin.Application) {
	lCLI := app.Command("layout", "Commands for layout")

	clm := &layout{}
	cl := lCLI.Command("create", "Create Layout").Action(clm.layoutCreate)

	lCLI.Flag("id", "Name of the layout").Required().Short('l').StringVar(&clm.id)
	lCLI.Flag("workspace-id", "Workspace name").Required().Short('w').StringVar(&clm.workspaceId)
	cl.Flag("dir", "Absolute path of directory where layout files exist").Required().Short('d').StringVar(&clm.dirName)

	lCLI.Command("get", "Get Layout").Action(clm.layoutGet)

	al := lCLI.Command("apply", "Apply layout").Action(clm.layoutApply)
	dl := lCLI.Command("destroy", "Destroy layout").Action(clm.layoutDestroy)

	al.Flag("dry", "Dry apply for in memory plan").BoolVar(&clm.dry)
	al.Flag("vars", "Path of vars file.").Short('v').StringVar(&clm.varsPath)
	dl.Flag("vars", "Path of vars file.").Short('v').StringVar(&clm.varsPath)

	lCLI.Command("state", "Get layout's current state").Action(clm.layoutStateGet)
}
