package main

import (
	"encoding/json"
	"os"
)

var (
	FILEPATH = "./config.json"
)

// TODO: Improve model to be easier to manage and navigate too
type config struct {
	User  string `json:"user"`
	Repos []repo `json:"repos"`
}

type repo struct {
	Owner        string `json:"owner"`
	Repo         string `json:"repo"`
	Merge        bool   `json:"merge"`
	PullsToMerge []int  `json:"pulls"`
}

// Add PR to be merged
func (r *repo) AddPull(number int) {
	r.PullsToMerge = append(r.PullsToMerge, number)
}

// Remove PR to be merged
func (r *repo) RemovePull(key int) {
	if len(r.PullsToMerge) == 1 {
		r.PullsToMerge = []int{}
	} else {
		r.PullsToMerge = append(r.PullsToMerge[:key], r.PullsToMerge[key+1:]...)
	}
}

// Load configuration from a file
// TODO: Implement Go Viper to manage config more easily
func initConfig() (*config, error) {
	var config config

	raw, err := os.ReadFile(FILEPATH)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
