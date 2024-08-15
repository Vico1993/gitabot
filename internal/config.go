package main

import (
	"encoding/json"
	"os"
)

var (
	FILEPATH = "./config.json"
)

type repo struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Merge bool   `json:"merge"`
}

// Load configuration from a file
func initConfig() ([]repo, error) {
	var repositories []repo

	raw, err := os.ReadFile(FILEPATH)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(raw, &repositories)

	return repositories, nil
}
