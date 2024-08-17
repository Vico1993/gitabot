package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	service "github.com/Vico1993/gitabot/internal/services"
	"github.com/Vico1993/gitabot/internal/utils"
	"github.com/google/go-github/v63/github"
	"github.com/subosito/gotenv"
)

var (
	DEPENDABOT_LOGIN    = "dependabot[bot]"
	DEPENDABOT_TYPE     = "Bot"
	APPROVE             = "APPROVE"
	APPROVED            = "APPROVED"
	FILTER_LATEST       = "latest"
	WAIT_GROUP          sync.WaitGroup
	PR_APPROVED         = 0
	PR_MERGED           = 0
	PR_NEEDED_ATTENTION = []string{}
	PR_MERGED_ERROR     = 0
)

func main() {
	// load .env file if any otherwise use env set
	_ = gotenv.Load()

	// Load service
	err := service.Init()
	if err != nil {
		fmt.Println("Issue initialtion services")
		log.Fatalln(err)
	}

	config, err := initConfig()
	if err != nil {
		fmt.Println("Issue retriving config")
		log.Fatalln(err)
	}
	_ = service.Telegram.PostMessage("🤖 🚧 [Gitabot]: Starting...")

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
	for _, repository := range config.Repos {
		WAIT_GROUP.Add(1)
		repository := repository

		go func() {
			defer WAIT_GROUP.Done()

			err := fetchPullRequestsApprovals(client, &repository, config.User)
			if err != nil {
				log.Fatalln(err)
			}

			if repository.Merge && len(repository.PullsToMerge) >= 1 {
				err := fetchPullRequestsMerged(client, &repository, config.User)
				if err != nil {
					log.Fatalln(err)
				}
			}

			fmt.Println("Done: ", repository.Owner, repository.Repo)
			fmt.Println(utils.ToJson(repository))
		}()
	}

	WAIT_GROUP.Wait()

	if len(PR_NEEDED_ATTENTION) > 0 {
		_ = service.Telegram.PostMessage("🤖 🟥 [Gitabot]: Number of PR that need attention " + strconv.Itoa(len(PR_NEEDED_ATTENTION)))

		fmt.Println("Few pull requests need your attention")
		fmt.Println(utils.ToJson(PR_NEEDED_ATTENTION))
	}

	_ = service.Telegram.PostMessage("🤖 🟩 [Gitabot]: Number of PR approved " + strconv.Itoa(PR_APPROVED) + " \n\n🤖 🟪 [Gitabot]: Number of PR merged " + strconv.Itoa(PR_MERGED) + "\n\n🤖 ✅ [Gitabot]: Done")

	if PR_MERGED > 0 {
		fmt.Println("Pull Requests merged: ", PR_MERGED)
		fmt.Println("Pull Requests merged - Error: ", PR_MERGED_ERROR)
	}

	fmt.Println("Done!")
}

// Fetch pull request from Repository and approving the repo
func fetchPullRequestsApprovals(client *github.Client, repository *repo, user string) error {
	pulls, err := featchAllPages(func(pageNumber int) ([]*github.PullRequest, *github.Response, error) {
		return client.PullRequests.List(
			context.Background(),
			repository.Owner,
			repository.Repo,
			&github.PullRequestListOptions{
				State: "open",
				ListOptions: github.ListOptions{
					PerPage: 2,
					Page:    pageNumber,
				},
			})
	})
	if err != nil {
		fmt.Println("Impossible to retrieve Pull Requests")
		return err
	}

	for _, p := range pulls {
		// Filter none dependabot PR
		if p.User.GetType() != DEPENDABOT_TYPE || p.User.GetLogin() != DEPENDABOT_LOGIN || p.GetDraft() {
			continue
		}

		isMergeable, err := isWorkflowSuccessfull(
			client,
			repository.Owner,
			repository.Repo,
			p.Head.GetRef(),
		)
		if err != nil {
			fmt.Println("Enable to fetch pull request workflows")
			return err
		}

		pull, _, err := client.PullRequests.Get(
			context.Background(),
			repository.Owner,
			repository.Repo,
			p.GetNumber(),
		)
		if err != nil {
			fmt.Println("Enable to fetch pull request")
			return err
		}

		isApprovable, err := isPullsApprovable(
			client,
			repository.Owner,
			repository.Repo,
			user,
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

		// To be mergeable, all checks need to be good
		// And no conflict need to be detected
		if isMergeable && pull.GetMergeable() {
			_, _, err := client.PullRequests.CreateReview(
				context.TODO(),
				repository.Owner,
				repository.Repo,
				p.GetNumber(),
				&github.PullRequestReviewRequest{
					Event: &APPROVE,
				},
			)
			if err != nil {
				fmt.Println("Enable to Approve pull request")
				return err
			}

			PR_APPROVED += 1

			// Add to array
			if repository.Merge {
				repository.AddPull(p.GetNumber())
			}
		} else {
			PR_NEEDED_ATTENTION = append(
				PR_NEEDED_ATTENTION,
				p.GetHTMLURL(),
			)
		}
	}

	return nil
}

// Checl if we can approve the PR
func isPullsApprovable(client *github.Client, owner, repo, user string, pullNumber int, commit string) (bool, error) {
	reviews, _, err := client.PullRequests.ListReviews(
		context.Background(),
		owner,
		repo,
		pullNumber,
		&github.ListOptions{
			PerPage: 100,
		},
	)
	if err != nil {
		return false, err
	}

	for _, review := range reviews {
		// if user already approved the PR for this commit
		if review.GetCommitID() == commit && review.GetState() == APPROVED && review.User.GetLogin() == user {
			return false, nil
		}
	}

	return true, nil
}

// Parse all workflows for this branch
func isWorkflowSuccessfull(client *github.Client, owner, repo, branchName string) (bool, error) {
	checks, _, err := client.Checks.ListCheckRunsForRef(
		context.Background(),
		owner,
		repo,
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

// Parsing a Github response and going threw everything to return all
func featchAllPages[T any](f func(int) ([]T, *github.Response, error)) ([]T, error) {
	d := []T{}
	pageNumber := 1
	for {
		data, res, err := f(pageNumber)

		if err != nil {
			return d, err
		}

		d = append(
			d,
			data...,
		)

		if res.NextPage == 0 {
			break
		} else {
			pageNumber += 1
		}
	}

	return d, nil
}

// Will handle logic for merging Pull requests
func fetchPullRequestsMerged(client *github.Client, repository *repo, user string) error {
	prNumberToMerge := repository.PullsToMerge[0]

	_, _, err := client.PullRequests.Merge(
		context.Background(),
		repository.Owner,
		repository.Repo,
		prNumberToMerge,
		"chore(deps): Update dep from Dependabot",
		&github.PullRequestOptions{},
	)
	if err != nil {
		PR_MERGED_ERROR += 1
		fmt.Println("Enable to Merge pull request:", repository.Owner, repository.Repo, prNumberToMerge)
		return err
	}

	// Remove pr merged
	repository.RemovePull(0)

	// Request Rebase
	var body = "@dependabot rebase"
	for _, prNumber := range repository.PullsToMerge {
		_, _, err := client.Issues.CreateComment(
			context.Background(),
			repository.Owner,
			repository.Repo,
			prNumber,
			&github.IssueComment{
				Body: &body,
				User: &github.User{
					Login: &user,
				},
			},
		)

		if err != nil {
			fmt.Println("Enable to Comment pull request:", repository.Owner, repository.Repo, prNumber)
			return err
		}
	}

	return nil
}
