package main

import (
	"log"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)
var wID *string
var lID *string
var jID *string

func varsAdd(c *kingpin.ParseContext) error {
	log.Println(*wID)
	log.Println(*lID)
	log.Println(*jID)

	log.Println("Vars command: Create vars for workspace, layout or job.")
	return nil
}

func varsGet(c *kingpin.ParseContext) error {
	log.Println(*wID)
	log.Println(*lID)
	log.Println(*jID)

	log.Println("Vars command: Get vars for workspace, layout or job.")
	return nil
}

func addVarsCommand(app *kingpin.Application) {
	// Add subcommands for the command "vars"
	vApp := app.Command("vars", "Vars for workspace, layout and job.")
	wID = vApp.Flag("workspace_id", "Workspace ID").Short('w').String()
	lID = vApp.Flag("layout_id", "Layout ID").Short('l').String()
	jID = vApp.Flag("job_id", "Job ID").Short('j').String()

	vApp.Command("create", "Save the vars for the given dir tree [workspace, layout or job].").Action(varsAdd)
	vApp.Command("get", "Get vars for the given dir tree [workspace, layout or job].").Action(varsGet)
}
