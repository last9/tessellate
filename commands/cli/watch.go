package main

import (
	"log"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"github.com/tsocial/tessellate/server"
	"strings"
	"context"
)

type watch struct {
	workspaceID string
	layoutID string
	sURL string
	fURL string
}

func (w *watch) watchStart(c *kingpin.ParseContext) error {

	log.Println(w.workspaceID)
	log.Println(w.layoutID)
	log.Println(w.sURL)
	log.Println(w.fURL)

	log.Println("Watch command: Creating watch for given workspace and layout.")

	client := getClient()
	req := server.StartWatchRequest{WorkspaceId: strings.ToLower(w.workspaceID), Id: strings.ToLower(w.layoutID),
	SuccessCallback: w.sURL, FailureCallback: w.fURL }

	ctx := context.Background()
	if _, err := client.StartWatch(ctx, &req); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (w *watch) watchStop(c *kingpin.ParseContext) error {
	log.Println(w.workspaceID)
	log.Println(w.layoutID)
	log.Println(w.sURL)
	log.Println(w.fURL)

	log.Println("Watch command: Get vars for given layout's job.")

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
	wStart := wApp.Command("start", "Start the watch for the given workspace and layout," +
		" which triggers an http call on success or failure of the job.").Action(wCLI.watchStart)
	wStop := wApp.Command("stop", "Stop the watch for the given workspace and layout ID.").Action(wCLI.watchStop)

	// Tie flags to `watch start -w wID -l lID -s sURL -f fURL` command.
	wStart.Flag("workspace_id", "Workspace ID").Required().Short('w').StringVar(&wCLI.workspaceID)
	wStart.Flag("layout_id", "Layout ID").Required().Short('l').StringVar(&wCLI.layoutID)

	wStart.Flag("success_url", "URL to trigger on success of job.").Short('s').StringVar(&wCLI.sURL)
	wStart.Flag("failure_url", "URL to trigger on failure of job.").Short('f').StringVar(&wCLI.fURL)

	// Tie flags to `watch stop -w wID -l lID`
	wStop.Flag("workspace_id", "Workspace ID").Required().Short('w').StringVar(&wCLI.workspaceID)
	wStop.Flag("layout_id", "Layout ID").Required().Short('l').StringVar(&wCLI.layoutID)
}
