package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/Vico1993/gitabot/internal/utils"
	"github.com/google/go-github/v63/github"
)

const PER_PAGE = 100

var (
	FILTER_LATEST    = "latest"
	DEPENDABOT_LOGIN = "dependabot[bot]"
	DEPENDABOT_TYPE  = "Bot"
	APPROVE          = "APPROVE"
	APPROVED         = "APPROVED"
)

type repository struct {
	client       *github.Client
	owner        string
	name         string
	user         string
	shouldMerge  bool
	pullToMerge  int
	pullToRebase []int
}

// Build the Repository object
func InitRepository(client *github.Client, owner, name, user string, shouldMerge bool) *repository {
	return &repository{
		client:       client,
		owner:        owner,
		name:         name,
		user:         user,
		shouldMerge:  shouldMerge,
		pullToMerge:  0,
		pullToRebase: []int{},
	}
}

// Function that will fetch and parsed each dependabot pulls
// Will approved if every check are succesfull
func (r *repository) HandleDependabotPulls() error {
	pullsList, err := r.fetchPullRequests()
	if err != nil {
		return err
	}

	for _, p := range pullsList {
		// Filter none dependabot PR
		if p.User.GetType() != DEPENDABOT_TYPE || p.User.GetLogin() != DEPENDABOT_LOGIN || p.GetDraft() {
			continue
		}

		pull, _, err := r.client.PullRequests.Get(
			context.Background(),
			r.owner,
			r.name,
			p.GetNumber(),
		)
		if err != nil {
			fmt.Println("Enable to fetch pull request")
			return err
		}

		isApprovable, err := r.isPullApprovable(
			pull.GetNumber(),
			pull.Base.GetSHA(),
		)
		if err != nil {
			fmt.Println("Enable to fetch pull request reviews")
			return err
		}
		if !isApprovable {
			continue
		}

		isMergeable, err := r.isBranchChecksSuccessful(
			p.Head.GetRef(),
		)
		if err != nil {
			fmt.Println("Enable to fetch branch checks")
			return err
		}

		// To be mergeable, all checks need to be good
		// And no conflict need to be detected
		if isMergeable && pull.GetMergeable() {
			r.handleApproval(pull.GetNumber())
		} else {
			PR_NEEDED_ATTENTION = append(
				PR_NEEDED_ATTENTION,
				p.GetHTMLURL(),
			)
		}
	}

	return nil
}

// Fetch all github pull request for a repository
func (r *repository) fetchPullRequests() ([]*github.PullRequest, error) {
	pulls, err := utils.FetchPages(
		func(pageNumber int) ([]*github.PullRequest, *github.Response, error) {
			return r.client.PullRequests.List(
				context.Background(),
				r.owner,
				r.name,
				&github.PullRequestListOptions{
					State: "open",
					ListOptions: github.ListOptions{
						PerPage: PER_PAGE,
						Page:    pageNumber,
					},
				})
		},
	)
	if err != nil {
		fmt.Println("Impossible to retrieve Pull Requests")
		return nil, err
	}

	return pulls, nil
}

// Parse checks to be sure pull request is mergeable
func (r *repository) isBranchChecksSuccessful(branchName string) (bool, error) {
	checks, _, err := r.client.Checks.ListCheckRunsForRef(
		context.Background(),
		r.owner,
		r.name,
		branchName,
		&github.ListCheckRunsOptions{
			Filter: &FILTER_LATEST,
		},
	)
	if err != nil {
		return false, err
	}

	for _, run := range checks.CheckRuns {
		if run.GetConclusion() != "success" {
			return false, nil
		}
	}

	return true, nil
}

// Check if we can approve the PR
func (r *repository) isPullApprovable(pullNumber int, commit string) (bool, error) {
	reviews, _, err := r.client.PullRequests.ListReviews(
		context.Background(),
		r.owner,
		r.name,
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
		if review.GetCommitID() == commit && review.GetState() == APPROVED && review.User.GetLogin() == r.user {
			return false, nil
		}
	}

	return true, nil
}

// Will process the approval
func (r *repository) handleApproval(pullNumber int) error {
	_, _, err := r.client.PullRequests.CreateReview(
		context.TODO(),
		r.owner,
		r.name,
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

	if !r.shouldMerge {
		return nil
	}

	if r.pullToMerge == 0 {
		r.pullToMerge = pullNumber
	} else {
		r.pullToRebase = append(r.pullToRebase, pullNumber)
	}

	return nil
}

// Handle logic around merging Pull + rebasing next PR
func (r *repository) HandleMerge() error {
	if r.pullToMerge == 0 {
		return errors.New("No pull request to merge")
	}

	_, _, err := r.client.PullRequests.Merge(
		context.Background(),
		r.owner,
		r.name,
		r.pullToMerge,
		"chore(deps): Update dep from Dependabot",
		&github.PullRequestOptions{},
	)
	if err != nil {
		PR_MERGED_ERROR += 1
		fmt.Println("Enable to Merge pull request:", r.owner, r.name, r.pullToMerge)
		return err
	}

	// Request Rebase
	var body = "@dependabot rebase"
	for _, prNumber := range r.pullToRebase {
		_, _, err := r.client.Issues.CreateComment(
			context.Background(),
			r.owner,
			r.name,
			prNumber,
			&github.IssueComment{
				Body: &body,
				User: &github.User{
					Login: &r.user,
				},
			},
		)

		if err != nil {
			fmt.Println("Enable to Merge pull request:", r.owner, r.name, prNumber)
			return err
		}
	}

	return nil
}
