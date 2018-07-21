package main

import (
	"context"
	"log"
	"strings"

	"github.com/tsocial/tessellate/server"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var wid *string

func workspaceAdd(c *kingpin.ParseContext) error {
	client := getClient()
	req := server.SaveWorkspaceRequest{Id: strings.ToLower(*wid)}

	ctx := context.Background()
	if _, err := client.SaveWorkspace(ctx, &req); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func workspaceGet(c *kingpin.ParseContext) error {
	client := getClient()
	req := server.GetWorkspaceRequest{Id: strings.ToLower(*wid)}

	ctx := context.Background()
	w, err := client.GetWorkspace(ctx, &req)
	if err != nil {
		log.Println(err)
		return err
	}

	prettyPrint(w)
	return nil
}

func addWorkspaceCommand(app *kingpin.Application) {
	w := app.Command("workspace", "Workspace")
	w.Command("create", "Create a workspace").Action(workspaceAdd)
	w.Command("get", "Get a workspace").Action(workspaceGet)

	wid = w.Flag("workspace_id", "Workspace Id").Short('w').Required().String()
}
