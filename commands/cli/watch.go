package main

import (
	"log"

	"context"
	"strings"

	"github.com/tsocial/tessellate/server"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type watch struct {
	workspaceID string
	layoutID    string
	sURL        string
	fURL        string
}

func (w *watch) watchStart(c *kingpin.ParseContext) error {
	log.Println("watch start -w " + w.workspaceID + " -l " + w.layoutID + " -s " + w.sURL + " -f " + w.fURL)

	client := getClient()
	req := server.StartWatchRequest{WorkspaceId: strings.ToLower(w.workspaceID), Id: strings.ToLower(w.layoutID),
		SuccessCallback: w.sURL, FailureCallback: w.fURL}

	ctx := context.Background()
	if _, err := client.StartWatch(ctx, &req); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (w *watch) watchStop(c *kingpin.ParseContext) error {
	log.Println("watch stop -w " + w.workspaceID + " -l " + w.layoutID)

	client := getClient()
	req := server.StopWatchRequest{WorkspaceId: strings.ToLower(w.workspaceID), Id: strings.ToLower(w.layoutID)}

	ctx := context.Background()
	if _, err := client.StopWatch(ctx, &req); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func addWatchCommand(app *kingpin.Application) {
	// Add subcommands for the command "vars"
	wApp := app.Command("watch", "Watch for a job.")

	wCLI := &watch{}

	// Sub commands : start and stop.
	wStart := wApp.Command("start", "Start the watch for the given workspace and layout,"+
		" which triggers an http call on success or failure of the job.").Action(wCLI.watchStart)
	wApp.Command("stop", "Stop the watch for the given workspace and layout ID.").Action(wCLI.watchStop)

	// Tie flags to `watch start -w wID -l lID -s sURL -f fURL` command.
	wApp.Flag("workspace_id", "Workspace ID").Required().Short('w').StringVar(&wCLI.workspaceID)
	wApp.Flag("layout_id", "Layout ID").Required().Short('l').StringVar(&wCLI.layoutID)

	wStart.Arg("success_callback", "URL to trigger on success of job.").StringVar(&wCLI.sURL)
	wStart.Arg("failure_callback", "URL to trigger on failure of job.").StringVar(&wCLI.fURL)
}
