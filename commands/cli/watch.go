package main

import (
	"log"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var workspaceID *string
var layoutID *string
var sURL *string
var fURL *string

func watchStart(c *kingpin.ParseContext) error {
	log.Println(*workspaceID)
	log.Println(*layoutID)
	log.Println(*sURL)
	log.Println(*fURL)

	log.Println("Watch command: Create watch for given layout's job.")
	return nil
}

func watchStop(c *kingpin.ParseContext) error {
	log.Println(*workspaceID)
	log.Println(*layoutID)
	log.Println(*sURL)
	log.Println(*fURL)

	log.Println("Watch command: Get vars for given layout's job.")
	return nil
}

func addWatchCommand(app *kingpin.Application) {
	// Add subcommands for the command "vars"
	wApp := app.Command("watch", "Watch for a job.")

	wStart := wApp.Command("start", "Start the watch for the given workspace and layout, which triggers an http call on success or failure of the job.").Action(watchStart)
	wStop := wApp.Command("stop", "Stop the watch for the given workspace and layout ID.").Action(watchStop)

	workspaceID = wStart.Flag("workspace_id", "Workspace ID").Required().Short('w').String()
	layoutID = wStart.Flag("layout_id", "Layout ID").Required().Short('l').String()

	sURL = wStart.Flag("success_url", "URL to trigger on success of job.").Short('s').String()
	fURL = wStart.Flag("failure_url", "URL to trigger on failure of job.").Short('f').String()

	workspaceID = wStop.Flag("workspace_id", "Workspace ID").Required().Short('w').String()
	layoutID = wStop.Flag("layout_id", "Layout ID").Required().Short('l').String()
}
