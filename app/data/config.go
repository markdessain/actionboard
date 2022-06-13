package data

import "os"

type Config struct {
	GithubToken string
	Repositories []string
	StorePath string
	Port string
	Days int
	SleepInterval int
}

// NewConfig setups up a config and combines the arguments with environment variables to ensure
// all the correct values are set.
func NewConfig(githubToken string, repositories []string, storePath string, port string, days int, sleepInterval int) Config {

	// Try using an environment variable instead if missing.
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	}

	return Config{
		GithubToken:  githubToken,
		Repositories: repositories,
		StorePath:     storePath,
		Port:		  port,
		Days:         days,
		SleepInterval: sleepInterval,
	}
}
