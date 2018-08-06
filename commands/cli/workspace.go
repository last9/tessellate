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

	log.Println("----------------------------------------------")
	for _, w := range w.Workspaces {
		if w.Vars != nil && len(w.Vars) > 0 {
			ppWorkspace(w)
		} else {
			prettyPrint(w)
		}
		log.Println("----------------------------------------------")
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

func addWorkspaceCommand(app *kingpin.Application) {
	w := app.Command("workspace", "Workspace")

	wc := w.Command("create", "Create a workspace.").Action(workspaceAdd)
	wc.Flag("workspace_id", "Workspace Id").Short('w').Required().StringVar(&wid)
	wc.Flag("providers", "Path to providers.tf.json").Short('p').StringVar(&providerFilePath)

	wg := w.Command("get", "Get a workspace.").Action(workspaceGet)
	wg.Flag("workspace_id", "Workspace Id").Short('w').Required().StringVar(&wid)

	w.Command("list", "Get All Workspaces.").Action(workspaceAll)
}
