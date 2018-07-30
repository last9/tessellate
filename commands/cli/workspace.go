package main

import (
	"log"
	"strings"

	"io/ioutil"

	"github.com/tsocial/tessellate/server"
	"gopkg.in/alecthomas/kingpin.v2"
)

var wid *string
var providerFilePath *string

func workspaceAdd(_ *kingpin.ParseContext) error {
	client := getClient()
	req := server.SaveWorkspaceRequest{Id: strings.ToLower(*wid)}

	var err error
	if *providerFilePath != "" {
		req.Providers, err = ioutil.ReadFile(*providerFilePath)
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
	req := server.GetWorkspaceRequest{Id: strings.ToLower(*wid)}

	w, err := client.GetWorkspace(makeContext(nil), &req)
	if err != nil {
		log.Println(err)
		return err
	}

	prettyPrint(w)
	return nil
}

func addWorkspaceCommand(app *kingpin.Application) {
	w := app.Command("workspace", "Workspace")
	wc := w.Command("create", "Create a workspace").Action(workspaceAdd)
	w.Command("get", "Get a workspace").Action(workspaceGet)

	wid = w.Flag("workspace_id", "Workspace Id").Short('w').Required().String()
	providerFilePath = wc.Flag("providers", "Path to providers.tf.json").Short('p').String()
}
