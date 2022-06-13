package github

import (
	"context"
	"github.com/google/go-github/v44/github"
	"log"
	"time"
)

// SyncWorkflows will load all the pages and grab every workflow for the selected repository.
func SyncWorkflows(owner, repo string, client *github.Client, workflowChannel chan *github.Workflow) {
	ctx := context.Background()
	workflows := make(map[int64]*github.Workflow)
	opt := &github.ListOptions{PerPage: 100}

	for {
		runs, resp, err := client.Actions.ListWorkflows(ctx, owner, repo, opt)
		if err != nil {
			log.Println(err)
			close(workflowChannel)
			return
		}
		for _, w := range runs.Workflows {
			workflows[w.GetID()] = w
			workflowChannel <- w
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	close(workflowChannel)
}

// SyncWorkflowRuns will load all the pages until it reaches the time duration and grab the runs for the selected repository.
func SyncWorkflowRuns(owner, repo string, client *github.Client, historyDuration time.Duration, workflowRunChannel chan *github.WorkflowRun) {

	ctx := context.Background()
	now := time.Now().UTC()
	earliest := now.Add(historyDuration * -1)
	opt2 := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	done := false

	for {
		runs, resp, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opt2)
		if err != nil {
			log.Println(err)
			// If there is an error, stop checking, and it will retry on the next run
			close(workflowRunChannel)
			return
		}
		for _, r := range runs.WorkflowRuns {
			if r.GetCreatedAt().UTC().String() < earliest.String() {
				done = true
			} else {
				workflowRunChannel <- r
			}
		}
		if done {
			break
		}
		if resp.NextPage == 0 {
			break
		}
		opt2.Page = resp.NextPage
	}
	close(workflowRunChannel)
}

// ByWorkflowAndTime will order the runs first by the workflows and then by the run started at.
type ByWorkflowAndTime []*github.WorkflowRun

func (a ByWorkflowAndTime) Len() int {
	return len(a)

}
func (a ByWorkflowAndTime) Less(i, j int) bool {
	if a[i].GetWorkflowID() < a[j].GetWorkflowID() {
		return true
	} else if a[i].GetWorkflowID() > a[j].GetWorkflowID() {
		return false
	} else {
		return a[i].GetRunStartedAt().String() < a[j].GetRunStartedAt().String()
	}
}

func (a ByWorkflowAndTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
