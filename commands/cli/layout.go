package main

import (
	"log"

	"os"

	"path/filepath"

	"fmt"

	"context"
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

	var maps []interface{}

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

		maps = append(maps, fObj)
	}

	finalMap := mergeMaps(maps...)

	layoutBytes, err := json.Marshal(finalMap)
	if err != nil {
		log.Println(err)
		return err
	}

	req := &server.SaveLayoutRequest{
		Id:          cm.id,
		WorkspaceId: cm.workspaceId,
		Plan:        layoutBytes,
	}

	_, err = getClient().SaveLayout(context.Background(), req)

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

	resp, err := getClient().GetLayout(context.Background(), req)
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
		Dry:         false,
	}

	resp, err := getClient().ApplyLayout(context.Background(), req)
	if err != nil {
		return err
	}

	state := resp.Status

	prettyPrint(state)
	return nil
}

func (cm *layout) layoutDestroy(c *kingpin.ParseContext) error {
	vars, err := getVars(cm.varsPath)
	if err != nil {
		return err
	}

	req := &server.ApplyLayoutRequest{
		Id:          cm.id,
		WorkspaceId: cm.workspaceId,
		Vars:        vars,
		Dry:         false,
	}

	resp, err := getClient().DestroyLayout(context.Background(), req)
	if err != nil {
		return err
	}

	state := resp.Status

	prettyPrint(state)
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

	lCLI.Flag("id", "Name of the layout").Required().StringVar(&clm.id)
	lCLI.Flag("workspace-id", "Workspace name").Required().StringVar(&clm.workspaceId)
	cl.Flag("dir", "Absolute path of directory where layout files exist").Required().StringVar(&clm.dirName)

	lCLI.Command("get", "Get Layout").Action(clm.layoutGet)

	al := lCLI.Command("apply", "Apply layout").Action(clm.layoutApply)
	dl := lCLI.Command("destroy", "Destroy layout").Action(clm.layoutDestroy)

	al.Flag("vars", "Path of vars file.").StringVar(&clm.varsPath)
	dl.Flag("vars", "Path of vars file.").StringVar(&clm.varsPath)
}

func mergeMaps(maps ...interface{}) interface{} {
	if len(maps) == 1 {
		return maps[0]
	}

	if len(maps) == 2 {
		return merge(maps[0], maps[1])
	}

	merged := merge(maps[0], maps[1])
	return mergeMaps(append(maps[2:], merged)...)
}

func merge(x1, x2 interface{}) interface{} {

	switch x1 := x1.(type) {
	case map[string]interface{}:
		x2, ok := x2.(map[string]interface{})
		if !ok {
			return x1
		}
		for k, v2 := range x2 {
			if v1, ok := x1[k]; ok {
				x1[k] = merge(v1, v2)
			} else {
				x1[k] = v2
			}
		}
	case nil:
		// merge(nil, map[string]interface{...}) -> map[string]interface{...}
		x2, ok := x2.(map[string]interface{})
		if ok {
			return x2
		}
	}
	return x1
}
