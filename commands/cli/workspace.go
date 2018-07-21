package main

import (
	"log"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var wid *string

func workspaceAdd(c *kingpin.ParseContext) error {
	log.Println(*wid)
	log.Println("Workspace command add")
	return nil
}

func workspaceGet(c *kingpin.ParseContext) error {
	log.Println(*wid)
	log.Println("Workspace command get")
	return nil
}

func addWorkspaceCommand(app *kingpin.Application) {
	w := app.Command("workspace", "Workspace")
	wid = w.Flag("workspace_id", "Workspace Id").Short('w').String()
	w.Command("create", "Create a workspace").Action(workspaceAdd)
	w.Command("get", "Get a workspace").Action(workspaceGet)
}
