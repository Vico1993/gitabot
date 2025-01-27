package main

import (
	"context"
	"fmt"

	"github.com/Vico1993/gitabot/internal/utils"
	"github.com/google/go-github/v63/github"
)

// Find all dependabot Pull requests that I can have access
func findDependabotIssues(client *github.Client, name string) ([]*github.Issue, error) {
	query := "is:open is:pr author:" + DEPENDABOT_LOGIN

	issues := []*github.Issue{}

	// Finding dependabots pulls - Repo owner
	orgIssues, err := utils.FetchPages(
		func(pageNumber int) ([]*github.Issue, *github.Response, error) {
			result, res, err := client.Search.Issues(
				context.TODO(),
				query+" org:"+name,
				&github.SearchOptions{
					ListOptions: github.ListOptions{
						PerPage: PER_PAGE,
						Page:    pageNumber,
					},
				},
			)

			return result.Issues, res, err
		},
	)
	if err != nil {
		fmt.Println("Impossible to retrieve Issues for org:" + name)
		return nil, err
	}

	for _, orgIssue := range orgIssues {
		alreadyIn := false
		for _, issue := range issues {
			if issue.GetID() == orgIssue.GetID() {
				alreadyIn = true
				break
			}
		}

		if !alreadyIn {
			issues = append(issues, orgIssue)
		}
	}

	// Finding dependabots pulls - Review Request
	reviewIssues, err := utils.FetchPages(
		func(pageNumber int) ([]*github.Issue, *github.Response, error) {
			result, res, err := client.Search.Issues(
				context.TODO(),
				query+" review-requested:"+name,
				&github.SearchOptions{
					ListOptions: github.ListOptions{
						PerPage: PER_PAGE,
						Page:    pageNumber,
					},
				},
			)

			return result.Issues, res, err
		},
	)
	if err != nil {
		fmt.Println("Impossible to retrieve Issues for review-requested:" + name)
		return nil, err
	}

	for _, reviewIssue := range reviewIssues {
		alreadyIn := false
		for _, issue := range issues {
			if issue.GetID() == reviewIssue.GetID() {
				alreadyIn = true
				break
			}
		}

		if !alreadyIn {
			issues = append(issues, reviewIssue)
		}
	}

	return issues, nil
}
