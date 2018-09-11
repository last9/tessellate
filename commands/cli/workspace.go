package main

import (
	"log"
	"strings"

	"io/ioutil"

	"encoding/json"

	"github.com/tsocial/tessellate/server"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	wid              string
	providerFilePath string
)

func workspaceAdd(_ *kingpin.ParseContext) error {
	client := getClient()
	req := server.SaveWorkspaceRequest{Id: strings.ToLower(wid)}

	var err error
	if providerFilePath != "" {
		req.Providers, err = ioutil.ReadFile(providerFilePath)
		if err != nil {
			return err
		}
	} else {
		log.Println("warning: no provider file given")
	}

	if _, err := client.SaveWorkspace(makeContext(nil), &req); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func workspaceGet(_ *kingpin.ParseContext) error {
	client := getClient()
	req := server.GetWorkspaceRequest{Id: strings.ToLower(wid)}

	w, err := client.GetWorkspace(makeContext(nil), &req)
	if err != nil {
		log.Println(err)
		return err
	}

	if w.Vars == nil || len(w.Vars) == 0 {
		prettyPrint(w)
		return nil
	}

	ppWorkspace(w)
	return nil
}

func workspaceAll(_ *kingpin.ParseContext) error {
	client := getClient()
	req := server.Ok{}

	w, err := client.GetAllWorkspaces(makeContext(nil), &req)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, w := range w.Workspaces {
		out := workspaceMap(w)
		prettyPrint(out)
	}

	return nil
}

func ppWorkspace(w *server.Workspace) {
	var vars map[string]interface{}

	if err := json.Unmarshal(w.Vars, &vars); err != nil {
		log.Println(err)
		return
	}

	out := make(map[string]interface{})
	out["Name"] = w.Name
	out["Versions"] = w.Versions
	out["Version"] = w.Version
	out["Vars"] = vars
	prettyPrint(out)
}

func workspaceMap(w *server.Workspace) map[string]interface{} {
	type Var struct {
		Provider []map[string]struct {
			Access_key string
			Region     string
			Secret_key string
			Version    string
		}
		Variable struct {
			Region struct{}
		}
	}
	var vars Var
	out := make(map[string]interface{})
	if w.Vars != nil && len(w.Vars) > 0 {
		if err := json.Unmarshal(w.Vars, &vars); err != nil {
			log.Println(err)
			return nil
		}

		providers := ""
		for key := range vars.Provider {
			for k := range vars.Provider[key] {
				providers = providers + k + ", "
			}
		}
		if len(providers) > 0 {
			out["Providers"] = providers[:len(providers)-2]
		}
	}

	out["Name"] = w.Name
	out["Version"] = w.Version

	return out
}

func addWorkspaceCommand(app *kingpin.Application) {
	w := app.Command("workspace", "Workspace")

	wc := w.Command("create", "Create a workspace.").Action(workspaceAdd)
	wc.Flag("workspace_id", "Workspace Id").Short('w').Required().StringVar(&wid)
	wc.Flag("providers", "Path to providers.tf.json").Short('p').StringVar(&providerFilePath)

	wg := w.Command("get", "Get a workspace.").Action(workspaceGet)
	wg.Flag("workspace_id", "Workspace Id").Short('w').Required().StringVar(&wid)

	w.Command("list", "Get All Workspaces.").Action(workspaceAll)
}
