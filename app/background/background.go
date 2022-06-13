package background

import (
	"actionboard/app/data"
	githubl "actionboard/app/github"
	"errors"
	"github.com/google/go-github/v44/github"
	"log"
	"strings"
	"time"
)

// Start begins the background job. It will handle onExit
func Start(data *data.Data, config data.Config) {
	client := githubl.GetClient(data, config.GithubToken)

	go workflows(data, config, client)
	go workflowRuns(data, config, client)
}

// workflowRuns will sync the data every X seconds.
// each internal it will recheck the previous 24 hours for changes
// each day it will go back and recheck the previous Y days for changes
//
// Syncing can be paused and resumed if required
func workflowRuns(data *data.Data, config data.Config, client *github.Client) {

	durationMax := time.Hour * 24 * time.Duration(config.Days)
	durationDay := time.Hour * 24
	var duration time.Duration

	var t1, t2 time.Time
	t2 = data.LastRun

	for {

		if !data.Paused {
			t1 = time.Now().UTC()

			if t2.Year() == 1 {
				duration = durationMax
			} else if t1.Day() != t2.Day() {
				duration = durationMax
			} else {
				duration = durationDay
			}

			t2 = time.Now().UTC()

			for _, x := range config.Repositories {

				details := strings.Split(x, "/")

				if len(details) != 2 {
					log.Println(errors.New("repo should be in format <owner>/<repo>"))
				} else {

					workflowRunChannel := make(chan *github.WorkflowRun)
					go githubl.SyncWorkflowRuns(details[0], details[1], client, duration, workflowRunChannel)

					for v := range workflowRunChannel {
						data.WorkflowRuns[v.GetID()] = v

						data.Workflows[v.GetWorkflowID()] = &github.Workflow{
							ID:        v.WorkflowID,
							NodeID:    nil,
							Name:      v.Name,
							Path:      nil,
							State:     nil,
							CreatedAt: nil,
							UpdatedAt: nil,
							URL:       v.WorkflowURL,
							HTMLURL:   nil,
							BadgeURL:  nil,
						}
					}

				}

			}

			removeOldRuns(data, durationMax)
		}

		time.Sleep(time.Second * time.Duration(config.SleepInterval))
	}

}

func removeOldRuns(data *data.Data, durationMax time.Duration) {
	oldWorkflows := []int64{}
	for k, v := range data.WorkflowRuns {
		now := time.Now().UTC()
		earliest := now.Add(durationMax * -1)
		if v.GetCreatedAt().UTC().String() < earliest.String() {
			oldWorkflows = append(oldWorkflows, k)
		}
	}

	if len(oldWorkflows) > 0 {
		log.Println("Cleaning up:")
		for _, w := range oldWorkflows {
			log.Println("Delete workflow run:")
			log.Println(w)
			delete(data.WorkflowRuns, w)
		}
	}
}

func workflows(data *data.Data, config data.Config, client *github.Client) {

	if !data.Paused {
		for _, x := range config.Repositories {

			details := strings.Split(x, "/")

			workflowChannel := make(chan *github.Workflow)
			go githubl.SyncWorkflows(details[0], details[1], client, workflowChannel)

			for v := range workflowChannel {
				data.Workflows[v.GetID()] = &github.Workflow{
					ID:        v.ID,
					NodeID:    nil,
					Name:      v.Name,
					Path:      nil,
					State:     nil,
					CreatedAt: nil,
					UpdatedAt: nil,
					URL:       v.URL,
					HTMLURL:   nil,
					BadgeURL:  nil,
				}
			}
		}
	}
}