package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemovePullWithOneElement(t *testing.T) {
	repository := &repo{
		Owner:        "Owner",
		Repo:         "Repo",
		Merge:        true,
		PullsToMerge: []int{2},
	}

	repository.RemovePull(0)

	assert.Len(t, repository.PullsToMerge, 0, "Should be empty")
}

func TestAddPullWithOneElement(t *testing.T) {
	repository := &repo{
		Owner:        "Owner",
		Repo:         "Repo",
		Merge:        true,
		PullsToMerge: []int{},
	}

	repository.AddPull(2)

	assert.Len(t, repository.PullsToMerge, 1, "Should have 1 element")
}
