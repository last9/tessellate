package main

import (
	"log"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func workspaceCmd(c *kingpin.ParseContext) error {
	log.Println("Workspace commands")
	return nil
}

func addWorkspaceCommand(app *kingpin.Application) {
	w := app.Command("workspace", "Workspace").Action(workspaceCmd)
	w.Command("create", "Create a workspace")
	w.Command("get", "Get a workspace")
}
