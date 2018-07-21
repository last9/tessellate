package main

import (
	"log"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func watchCmd(c *kingpin.ParseContext) error {
	log.Println("Watch commands")
	return nil
}

func addWatchCommand(app *kingpin.Application) {
	vApp := app.Command("watch", "Watch for a job.").Action(watchCmd)
	vApp.Command("save", "Save the watch for the given job to trigger an http call on success or failure of the job.")
	vApp.Command("get", "Get the watch for the given job ID.")
}
