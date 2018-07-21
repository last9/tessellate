package main

import (
	"log"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func varsCmd(c *kingpin.ParseContext) error {
	log.Println("Vars commands")
	return nil
}

func addVarsCommand(app *kingpin.Application) {
	vApp := app.Command("vars", "Vars for workspace, layout and job.").Action(varsCmd)
	vApp.Command("save", "Save the vars for the given dir tree [workspace, layout or job].")
	vApp.Command("get", "Get vars for the given dir tree [workspace, layout or job].")
}
