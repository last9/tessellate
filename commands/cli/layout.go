package main

import (
	"bufio"
	"io"
	"log"
	"path"

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

func defaultBlackList() []string {
	return []string{"tfvars"}
}

func defaultManifest() []string {
	return []string{
		".tf.json",
	}
}

func readFileLines(file string) ([]string, error) {
	lines := []string{}

	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		lines = append(lines, string(line))
	}
	return lines, err
}

// candidateFiles matches files that should be uploaed or not
func candidateFiles(dirname string, manifest []string) ([]string, error) {
	if manifest == nil {
		manifest = defaultManifest()
	}

	blacklist := defaultBlackList()
	var files []string
	if err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		for _, b := range blacklist {
			if strings.Contains(path, b) {
				log.Printf("skipping %s", path)
				return nil
			}
		}

		for _, m := range manifest {
			if strings.HasSuffix(path, m) {
				files = append(files, path)
				break
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func (cm *layout) layoutCreate(c *kingpin.ParseContext) error {
	if _, err := os.Stat(cm.dirName); err != nil {
		log.Printf("Directory '%s' does not exist\n", cm.dirName)
	}

	manifest, err := readFileLines(path.Join(cm.dirName, ".tsl8"))
	if err != nil {
		log.Println(err)
	}

	files, err := candidateFiles(cm.dirName, manifest)
	if err != nil {
		return errors.Wrap(err, "Cannot get files")
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

func (cm *layout) layoutGetOutput(_ *kingpin.ParseContext) error {
	req := &server.GetOutputRequest{
		LayoutId:    cm.id,
		WorkspaceId: cm.workspaceId,
	}

	resp, err := getClient().GetOutput(makeContext(nil), req)
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

func addLayoutCommands(app *kingpin.Application) {
	lCLI := app.Command("layout", "Commands for layout")

	clm := &layout{}
	cl := lCLI.Command("create", "Create Layout").Action(clm.layoutCreate)

	lCLI.Flag("layout_id", "Name of the layout").Required().Short('l').StringVar(&clm.id)
	lCLI.Flag("workspace_id", "Workspace name").Required().Short('w').StringVar(&clm.workspaceId)
	cl.Flag("dir", "Absolute path of directory where layout files exist").Required().Short('d').StringVar(&clm.dirName)

	lCLI.Command("get", "Get Layout").Action(clm.layoutGet)

	al := lCLI.Command("apply", "Apply layout").Action(clm.layoutApply)
	dl := lCLI.Command("destroy", "Destroy layout").Action(clm.layoutDestroy)

	al.Flag("dry", "Dry apply for in memory plan").BoolVar(&clm.dry)
	al.Flag("vars", "Path of vars file.").Short('v').StringVar(&clm.varsPath)
	dl.Flag("vars", "Path of vars file.").Short('v').StringVar(&clm.varsPath)

	lCLI.Command("state", "Get layout's current state").Action(clm.layoutStateGet)
	lCLI.Command("output", "Get layout's output if exist").Action(clm.layoutGetOutput)
}
