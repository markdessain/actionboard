package data

import (
	"actionboard/app/http"
	"encoding/gob"
	"fmt"
	"github.com/google/go-github/v44/github"
	"log"
	"os"
	"time"
)

// Data holds all the application state in memory. It can be serialised to disk on
// an exit call.
type Data struct {
	Workflows map[int64]*github.Workflow
	WorkflowRuns map[int64]*github.WorkflowRun
	HttpCache http.MemoryCache
	LastRun time.Time
	Paused bool
}

func NewData(config Config) *Data {

	d, err := LoadData(config)

	// Loads the serialised data if it is available, otherwise create an empty state.
	if err != nil {
		return &Data{
			Workflows: make(map[int64]*github.Workflow),
			WorkflowRuns: make(map[int64]*github.WorkflowRun),
			HttpCache: http.NewMemoryCache(),
			Paused: false,
		}
	} else {
		return d
	}
}

// LoadData will load the file from disk and create a Data object.
func LoadData(config Config) (*Data, error) {

	file, err := os.Open(config.StorePath)
	if err != nil {
		return nil, err
	}

	d := Data{}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&d)
	file.Close()

	if err != nil {
		return nil, err
	}

	return &d, nil
}

// Save will serialise the data if the path exists
func (d Data) Save(config Config) error  {

	if config.StorePath == "" {
		return nil
	}

	log.Println("Saving")

	f, err := os.Create(config.StorePath)
	if err != nil {
		fmt.Println(err)
	}

	enc := gob.NewEncoder(f)
	err = enc.Encode(d)
	if err != nil {
		fmt.Println(err)
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
