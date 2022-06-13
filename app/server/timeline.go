package server

import (
	githubl "actionboard/app/github"
	"github.com/google/go-github/v44/github"
	"sort"
	"time"
)

type Cell struct {
	ColumnStart int
	ColumnEnd int
	RowStart int
	RowEnd int
	Value string
	State string
}

// generateTimeline will build up a timeline for all the workflow actions for a given day YYYY-MM-DD.
// It will ensure that no runs overlap with each other
// Workflow | 1 | 2 | 3 | 4 |
// A        | a | a |   | d |
// A        |   |   | b |   |
// A        |   |   | c | c |
//
// Each column will be split into a 15 minute interval - 360 per day.
// The logic to convert from a time xH yM to the index is:
// (hour*4)+(minute/15) = index
//
// The logic to convert from the index to a time xH yM is:
// index/4 = hour
// (index%4)*15 = minute
//
// Examples:
// 0hour 0min = 0
// 0hour 15min = 1
// 0hour 30min = 2
// 0hour 45min = 3
// 1hour 0min = 4
// 1 hour 15min = 5
// 1 hour 30min = 6
// 1 hour 45min = 7
// 2 hour 0min = 8
//
// TODO: Just looks at the current starting time. This won't work when jobs can span multiple 15 minutes
// TODO: Some duplicated logic
//
func generateTimeline(allWorkflowRuns map[int64]*github.WorkflowRun, date time.Time, zone *time.Location) []Cell {

	workflowRuns := []*github.WorkflowRun{}

	for _, v := range allWorkflowRuns {
		if v.GetRunStartedAt().In(zone).Year() == date.Year() && v.GetRunStartedAt().In(zone).Month() == date.Month() && v.GetRunStartedAt().In(zone).Day() == date.Day() {
			workflowRuns = append(workflowRuns, v)
		}
	}

	sort.Sort(githubl.ByWorkflowAndTime(workflowRuns))

	var currentWorkflowBlocks map[int][]*github.WorkflowRun
	currentWorkflowLength := 1
	nextWorkflowRow := 2

	currentWorkflowBlocks = make(map[int][]*github.WorkflowRun)
	var currentWorkflowId *int64
	var currentName *string

	var cells []Cell

	for _, v := range workflowRuns {

		if currentWorkflowId != nil && *currentWorkflowId != *v.WorkflowID {

			cells = append(cells, Cell{
				ColumnStart: 1,
				ColumnEnd:   5,
				RowStart:    nextWorkflowRow,
				RowEnd:      nextWorkflowRow+currentWorkflowLength,
				Value:       *currentName,
				State: 		 "label",
			})

			for k, w := range currentWorkflowBlocks {

				row := nextWorkflowRow
				for _, v2 := range w {
					cells = append(cells, Cell{
						ColumnStart: k,
						ColumnEnd:   k+1,
						RowStart:    row,
						RowEnd:      row+1,
						Value:       "", //v2.GetRunStartedAt().In(zone).String(),
						State: 		 v2.GetStatus() + "-" + v2.GetConclusion(),
					})

					row += 1
				}
			}

			nextWorkflowRow = nextWorkflowRow + currentWorkflowLength
			currentWorkflowLength = 1
			currentWorkflowBlocks = make(map[int][]*github.WorkflowRun)
		}
		currentWorkflowId = v.WorkflowID
		currentName = v.Name

		hour := v.GetRunStartedAt().In(zone).Hour()
		minute := v.GetRunStartedAt().In(zone).Minute()
		var nearestQuaterMinute int
		if minute < 15 {
			nearestQuaterMinute = 0
		} else if minute < 30 {
			nearestQuaterMinute = 15
		} else if minute < 45 {
			nearestQuaterMinute = 30
		} else {
			nearestQuaterMinute = 45
		}

		timeIndex := (hour*4)+(nearestQuaterMinute/15) + 1 + 1

		currentBlock := currentWorkflowBlocks[timeIndex]
		currentBlock = append(currentBlock, v)

		// Keep track of the max length, this could be caculated from list
		if len(currentBlock) > currentWorkflowLength {
			currentWorkflowLength = len(currentBlock)
		}

		// +1 as the first column contains the labels
		currentWorkflowBlocks[timeIndex] = currentBlock
	}

	if currentName != nil {

		// TODO: Duplicated from inside the loop
		cells = append(cells, Cell{
			ColumnStart: 1,
			ColumnEnd:   5,
			RowStart:    nextWorkflowRow ,
			RowEnd:      nextWorkflowRow + currentWorkflowLength ,
			Value:       *currentName,
			State:       "label",
		})

		for k, v := range currentWorkflowBlocks {
			row := nextWorkflowRow
			for _, v2 := range v {
				cells = append(cells, Cell{
					ColumnStart: k,
					ColumnEnd:   k+1,
					RowStart:    row,
					RowEnd:      row+1,
					Value:       "", //v2.GetRunStartedAt().In(zone).String(),
					State: 		 v2.GetStatus() + "-" + v2.GetConclusion(),
				})

				row += 1
			}
		}
	}

	return cells
}