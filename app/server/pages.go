package server

import (
	"actionboard/app/data"
	"html/template"
	"log"
	"net/http"
	"time"
)

// handleIndex shows the timeline for the selected filters
func handleIndex(data2 *data.Data, config data.Config) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var date time.Time
		var zone *time.Location
		var err error

		selectedTimezone := r.URL.Query()["timezone"]
		selectedDate := r.URL.Query()["date"]

		if len(selectedTimezone) == 1 {
			zone, err = time.LoadLocation(selectedTimezone[0])
			if err != nil {
				log.Println(err)
				zone = time.UTC
			}
		} else {
			zone = time.UTC
		}

		if len(selectedDate) == 1 {
			date, err = time.Parse("2006-01-02", selectedDate[0])
			if err != nil {
				log.Println(err)
				date = time.Now().UTC().In(zone)
			}
		} else {
			date = time.Now().UTC().In(zone)
		}
		repos := []Select{}

		for _, r := range config.Repositories {
			repos = append(repos, Select{
				Value:  r,
				Active: true,
			})
		}

		dates := []Select{}

		currentTime := time.Now().In(zone)

		for i := 0; i < config.Days; i++ {
			d := currentTime.Add(time.Hour * 24 * time.Duration(-i))
			dates = append(dates, Select{
				Value: d.Format("2006-01-02"),
				Active: date.Format("2006-01-02") == d.Format("2006-01-02"),
			})

			d = d.Add(time.Hour * 24 * -1)
		}

		b := struct {
			Data *data.Data
			Cells []Cell
			Zones []Select
			Dates []Select
			Repositories []Select
		}{
			Data:  data2,
			Cells: generateTimeline(data2.WorkflowRuns, date, zone),
			Zones: []Select{
				{"UTC", zone.String() == "UTC"},
				{"Europe/London", zone.String() == "Europe/London"},
				{"Europe/Copenhagen", zone.String()== "Europe/Copenhagen"},
			},
			Dates: dates,
			Repositories: repos,
		}

		t := template.Must(template.New("").Funcs(htmlFunctions()).ParseFS(fileSystem(), "html/base.html", "html/home.html"))
		t.ExecuteTemplate(w, "main", b)

	})
}

// handlePause will toggle the background syncing
func handlePause(data *data.Data, config data.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data.Paused = !data.Paused
		http.Redirect(w, r, "/", 301)
	})
}