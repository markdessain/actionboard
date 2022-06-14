package main

import (
	"actionboard/app/background"
	"actionboard/app/data"
	"actionboard/app/server"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Inputs that are provided on the commandline
var (
	githubToken string
	storePath string
	repositories []string
	port string
	days int
	sleepInterval int
)

// main is the single entry point that will run the web server and background syncing process.
// If multiple instances are run at the same time they will each work independently.
func main() {

	// Single command
	cmd := &cobra.Command{
		Use:   "",
		Run: func(cmd *cobra.Command, args []string) {
			// Load Configuration
			config := data.NewConfig(githubToken, repositories, storePath, port, days, sleepInterval)

			// Store the data in memory
			store := data.NewData(config)

			// Query the GitHub API in the background
			go background.Start(store, config)

			// Serve web requests.
			go server.Start(store, config)

			handleExit(store, config)
		},
	}

	// Load inputs from the user
	cmd.Flags().StringVarP(&githubToken, "github-token", "t", "", "The GitHub token to authenticate with the GitHub API.")
	cmd.Flags().StringVarP(&storePath, "store-path", "s", "", "The optional path to store the cached data on exit, once it restarts it will resume with this data.")
	cmd.Flags().StringVarP(&port, "port", "p", "8080", "The server port.")
	cmd.Flags().IntVarP(&days, "days", "d", 7, "How many days of data to load into memory.")
	cmd.Flags().IntVarP(&sleepInterval, "sleepInterval", "i", 100, "Number of seconds to sleep for before re-syncing with Github.")
	cmd.Flags().StringArrayVarP(&repositories, "repository", "r", []string{}, "Name of the repository in the format '<OWNER>/<REPOSITORY>'")

	// Run the command
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}


// handleExit will make sure that the data is serialised correctly once the application
// exits.
func handleExit(data *data.Data, config data.Config) {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	err := data.Save(config)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
