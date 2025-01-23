package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v63/github"
)

type Repository struct {
	client       *github.Client
	owner        string
	name         string
	user         string
	Pulls        []int
	pullToMerge  int
	pullToRebase []int
}

func InitRepository(client *github.Client, owner, name, user string, pullNumbers []int) *Repository {
	return &Repository{
		client:       client,
		owner:        owner,
		name:         name,
		user:         user,
		Pulls:        pullNumbers,
		pullToMerge:  0,
		pullToRebase: []int{},
	}
}

// Add a pull to the list
func (c *Repository) AddPulls(number int) {
	c.Pulls = append(c.Pulls, number)
}

func (c *Repository) HandlePulls() error {
	for _, pullNumber := range c.Pulls {
		pull, _, err := c.client.PullRequests.Get(
			context.Background(),
			c.owner,
			c.name,
			pullNumber,
		)
		if err != nil {
			fmt.Println("🔴: Enable to fetch pull request")
			return err
		}

		isApprovable, err := c.isPullApprovable(
			pull.GetNumber(),
			pull.Head.GetSHA(),
		)
		if err != nil {
			fmt.Println("🔴: Enable to fetch pull request reviews")
			return err
		}
		if !isApprovable {
			PR_ALREADY_APPROVED += 1
			continue
		}

		isMergeable, err := c.isBranchChecksSuccessful(
			pull.Head.GetRef(),
		)
		if err != nil {
			fmt.Println("🔴: Enable to fetch branch checks")
			return err
		}

		// To be mergeable, all checks need to be good
		// And no conflict need to be detected
		if isMergeable && pull.GetMergeable() {
			_, _, err := c.client.PullRequests.CreateReview(
				context.TODO(),
				c.owner,
				c.name,
				pullNumber,
				&github.PullRequestReviewRequest{
					Event: &APPROVE,
				},
			)
			if err != nil {
				fmt.Println("Enable to Approve pull request")
				return err
			}

			// Update stats
			PR_APPROVED += 1

			if strings.EqualFold(c.owner, c.user) {
				return nil
			}

			if c.pullToMerge == 0 {
				c.pullToMerge = pullNumber
			} else {
				c.pullToRebase = append(c.pullToRebase, pullNumber)
			}
		} else {
			PR_NEEDED_ATTENTION = append(
				PR_NEEDED_ATTENTION,
				pull.GetHTMLURL(),
			)
		}
	}

	return nil
}

// Parse checks to be sure pull request is mergeable
func (c *Repository) isBranchChecksSuccessful(branchName string) (bool, error) {
	checks, _, err := c.client.Checks.ListCheckRunsForRef(
		context.Background(),
		c.owner,
		c.name,
		branchName,
		&github.ListCheckRunsOptions{
			Filter: &FILTER_LATEST,
		},
	)
	if err != nil {
		return false, err
	}

	for _, run := range checks.CheckRuns {
		if run.GetConclusion() != "success" && run.GetConclusion() != "skipped" {
			return false, nil
		}
	}

	return true, nil
}

// Check if we can approve the PR
func (c *Repository) isPullApprovable(pullNumber int, commit string) (bool, error) {
	reviews, _, err := c.client.PullRequests.ListReviews(
		context.Background(),
		c.owner,
		c.name,
		pullNumber,
		&github.ListOptions{
			PerPage: PER_PAGE,
		},
	)
	if err != nil {
		return false, err
	}

	for _, review := range reviews {
		// if user already approved the PR for this commit
		if review.GetCommitID() == commit && review.GetState() == APPROVED && review.User.GetLogin() == c.user {
			fmt.Println("✅: user already approved the PR for this commit", c.owner,
				c.name,
				pullNumber,
			)
			return false, nil
		}
	}

	return true, nil
}

// Handle logic around merging Pull + rebasing next PR
func (c *Repository) HandleMerge() error {
	if c.pullToMerge == 0 {
		return errors.New("No pull request to merge")
	}

	_, _, err := c.client.PullRequests.Merge(
		context.Background(),
		c.owner,
		c.name,
		c.pullToMerge,
		"chore(deps): Update dep from Dependabot",
		&github.PullRequestOptions{},
	)
	if err != nil {
		PR_MERGED_ERROR += 1
		fmt.Println("Enable to Merge pull request:", c.owner, c.name, c.pullToMerge)
		return err
	} else {
		PR_MERGED += 1
	}

	// Request Rebase
	var body = "@dependabot rebase"
	for _, prNumber := range c.pullToRebase {
		_, _, err := c.client.Issues.CreateComment(
			context.Background(),
			c.owner,
			c.name,
			prNumber,
			&github.IssueComment{
				Body: &body,
				User: &github.User{
					Login: &c.user,
				},
			},
		)
		if err != nil {
			fmt.Println("Enable to Merge pull request:", c.owner, c.name, prNumber)
			return err
		}

		fmt.Println("✅ - PR Commented ", body, c.owner, c.name, prNumber)
	}

	return nil
}

// Log the information in the terminal
func (c *Repository) Logging(txt string) {
	fmt.Println(txt, c.owner, c.name)
}
