package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/commands/commons"
	"github.com/tsocial/tessellate/server"
	"gopkg.in/alecthomas/kingpin.v2"
)

const defaultAttempts = "3"

type layout struct {
	id          string
	workspaceId string
	dirName     string
	dry         bool
	varsPath    string
	retry       int64
}

func twoFAKey(args ...string) string {
	return strings.Join(args, "-")
}

func (cm *layout) layoutCreate(c *kingpin.ParseContext) error {
	if _, err := os.Stat(cm.dirName); err != nil {
		log.Printf("Directory '%s' does not exist\n", cm.dirName)
	}

	manifest, rErr := commons.ReadFileLines(path.Join(cm.dirName, ".tsl8"))
	if rErr != nil {
		log.Println("No manifest file found, moving ahead...")
	}

	files, cErr := commons.CandidateFiles(cm.dirName, manifest)
	if cErr != nil {
		return errors.Wrap(cErr, "Cannot get files")
	}

	if len(files) == 0 {
		return fmt.Errorf("no candidate files found in directory %s", cm.dirName)
	}

	wreq := server.GetWorkspaceRequest{Id: strings.ToLower(cm.workspaceId)}

	_, err := getClient().GetWorkspace(makeContext(nil, nil), &wreq)

	if err != nil {
		log.Println(fmt.Sprintf("Workspace %v does not exist", cm.workspaceId))
		return err
	}

	fLayout := map[string]interface{}{}

	// Will contain all candidate files.
	for _, f := range files {
		fBytes, fErr := ioutil.ReadFile(f)
		if fErr != nil {
			log.Println(fErr)
			return fErr
		}

		var fObj interface{}
		// If json, unmarshal as a JSON.
		if filepath.Ext(f) == ".json" {
			if err := json.Unmarshal(fBytes, &fObj); err != nil {
				log.Printf("invald json file: %s", f)
				return err
			}
		} else {
			// Copy the contents as they are. Say if the file is a .tmpl
			fObj = string(fBytes)
		}

		splits := strings.Split(f, "/")
		fLayout[splits[len(splits)-1]] = fObj
	}

	layoutBytes, mErr := json.Marshal(fLayout)
	if mErr != nil {
		log.Println(mErr)
		return mErr
	}

	req := &server.SaveLayoutRequest{
		Id:          cm.id,
		WorkspaceId: cm.workspaceId,
		Plan:        layoutBytes,
		Dry:         cm.dry,
	}
	resp, err := getClient().SaveLayout(makeContext(nil, NewTwoFA(twoFAKey(cm.workspaceId, cm.id), *codes)), req)
	if err != nil {
		log.Println(err)
		return err
	}
	prettyPrint("Layout created sucessfully. ID: " + resp.LayoutId)
	return nil
}

func (cm *layout) layoutGet(c *kingpin.ParseContext) error {
	req := &server.LayoutRequest{
		Id:          cm.id,
		WorkspaceId: cm.workspaceId,
	}

	resp, err := getClient().GetLayout(makeContext(nil, nil), req)
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
		Retry:       cm.retry,
	}

	resp, err := getClient().ApplyLayout(makeContext(nil, NewTwoFA(twoFAKey(cm.workspaceId, cm.id), *codes)), req)
	if err != nil {
		return err
	}

	prettyPrint("Check the link for job status: " + resp.Id)
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
		Retry:       cm.retry,
	}

	resp, err := getClient().DestroyLayout(makeContext(nil, NewTwoFA(twoFAKey(cm.workspaceId, cm.id), *codes)), req)
	if err != nil {
		return err
	}

	prettyPrint("Check the link for job status: " + resp.Id)
	return nil
}

func (cm *layout) layoutStateGet(_ *kingpin.ParseContext) error {
	req := &server.GetStateRequest{
		LayoutId:    cm.id,
		WorkspaceId: cm.workspaceId,
	}

	resp, err := getClient().GetState(makeContext(nil, NewTwoFA(twoFAKey(cm.workspaceId, cm.id), *codes)), req)
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

func (cm *layout) layoutGetOutput(_ *kingpin.ParseContext) error {
	req := &server.GetOutputRequest{
		LayoutId:    cm.id,
		WorkspaceId: cm.workspaceId,
	}

	resp, err := getClient().GetOutput(makeContext(nil, NewTwoFA(twoFAKey(cm.workspaceId, cm.id), *codes)), req)
	if err != nil {
		return err
	}

	var d interface{}
	if err := json.Unmarshal(resp.Output, &d); err != nil {
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

func (cm *layout) workspaceAllLayouts(_ *kingpin.ParseContext) error {
	client := getClient()
	req := server.GetWorkspaceLayoutsRequest{Id: cm.workspaceId}

	wL, err := client.GetWorkspaceLayouts(makeContext(nil, nil), &req)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, l := range wL.Layouts {
		prettyPrint(l.Id)
	}

	return nil
}

func addLayoutCommands(app *kingpin.Application) {
	lCLI := app.Command("layout", "Commands for layout")

	clm := &layout{}

	lCLI.Flag("workspace_id", "Workspace name").Required().Short('w').StringVar(&clm.workspaceId)
	lCLI.Command("list", "Get All Layouts.").Action(clm.workspaceAllLayouts)

	cl := lCLI.Command("create", "Create Layout").Action(clm.layoutCreate)
	cl.Flag("dry", "Saves a temporary layout for review").BoolVar(&clm.dry)
	cl.Flag("layout_id", "Name of the layout").Required().Short('l').StringVar(&clm.id)
	cl.Flag("dir", "Absolute path of directory where layout files exist").Required().Short('d').StringVar(&clm.dirName)

	gl := lCLI.Command("get", "Get Layout").Action(clm.layoutGet)
	gl.Flag("layout_id", "Name of the layout").Required().Short('l').StringVar(&clm.id)

	al := lCLI.Command("apply", "Apply layout").Action(clm.layoutApply)
	al.Flag("layout_id", "Name of the layout").Required().Short('l').StringVar(&clm.id)
	al.Flag("dry", "Dry apply for in memory plan").BoolVar(&clm.dry)
	al.Flag("retry", "Number of retries on layout apply, make it 0 for no retries, default is 3").
		Default(defaultAttempts).Int64Var(&clm.retry)
	al.Flag("vars", "Path of vars file.").Short('v').StringVar(&clm.varsPath)

	dl := lCLI.Command("destroy", "Destroy layout").Action(clm.layoutDestroy)
	dl.Flag("layout_id", "Name of the layout").Required().Short('l').StringVar(&clm.id)
	dl.Flag("vars", "Path of vars file.").Short('v').StringVar(&clm.varsPath)

	sl := lCLI.Command("state", "Get layout's current state").Action(clm.layoutStateGet)
	sl.Flag("layout_id", "Name of the layout").Required().Short('l').StringVar(&clm.id)

	ol := lCLI.Command("output", "Get layout's output if exist").Action(clm.layoutGetOutput)
	ol.Flag("layout_id", "Name of the layout").Required().Short('l').StringVar(&clm.id)
}
