package main

import (
	"encoding/json"
	"os"
)

var (
	FILEPATH = "./config.json"
)

type config struct {
	User  string `json:"user"`
	Repos []repo `json:"repos"`
}

type repo struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Merge bool   `json:"merge"`
}

// Load configuration from a file
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
