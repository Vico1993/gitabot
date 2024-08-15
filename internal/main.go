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
	WAIT_GROUP          sync.WaitGroup
	PR_APPROVED         = 0
	PR_MERGED           = 0
	PR_NEEDED_ATTENTION = []string{}
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
	service.Telegram.PostMessage("🤖 🚧 [Gitabot]: Starting...")

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
	for _, repo := range config.Repos {
		WAIT_GROUP.Add(1)
		repo := repo

		go func() {
			defer WAIT_GROUP.Done()

			err := fetchPullRequests(client, repo.Owner, repo.Repo, config.User)
			if err != nil {
				log.Fatalln(err)
			}
		}()
	}

	WAIT_GROUP.Wait()

	if len(PR_NEEDED_ATTENTION) > 0 {
		service.Telegram.PostMessage("🤖 🟥 [Gitabot]: Number of PR that need attention " + strconv.Itoa(len(PR_NEEDED_ATTENTION)))

		fmt.Println("Few pull requests need your attention")
		fmt.Println(utils.ToJson(PR_NEEDED_ATTENTION))
	}

	service.Telegram.PostMessage("🤖 🟩 [Gitabot]: Number of PR approved " + strconv.Itoa(PR_APPROVED) + " \n\n🤖 🟪 [Gitabot]: Number of PR merged " + strconv.Itoa(PR_MERGED) + "\n\n🤖 ✅ [Gitabot]: Done")
	fmt.Println("Done!")
}

// Fetch pull request from Repository and approving the repo
func fetchPullRequests(client *github.Client, owner, repo, user string) error {
	pulls, err := featchAllPages[*github.PullRequest](func(pageNumber int) ([]*github.PullRequest, *github.Response, error) {
		return client.PullRequests.List(
			context.Background(),
			owner,
			repo,
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

	for _, pull := range pulls {
		// Filter none dependabot PR
		if pull.User.GetType() != DEPENDABOT_TYPE || pull.User.GetLogin() != DEPENDABOT_LOGIN || pull.GetDraft() {
			continue
		}

		isMergeable, err := isWorkflowSuccessfull(client, owner, repo, pull.Head.GetRef())
		if err != nil {
			fmt.Println("Enable to fetch pull request workflows")
			return err
		}

		isApprovable, err := isPullsApprovable(client, owner, repo, user, pull.GetNumber())
		if err != nil {
			fmt.Println("Enable to fetch pull request reviews")
			return err
		}

		if !isApprovable {
			continue
		}

		if isMergeable {
			_, _, err := client.PullRequests.CreateReview(
				context.TODO(),
				owner,
				repo,
				pull.GetNumber(),
				&github.PullRequestReviewRequest{
					Event: &APPROVE,
				},
			)
			if err != nil {
				fmt.Println("Enable to Approve pull request")
				return err
			}

			PR_APPROVED += 1

			// TODO: Defined strategy for merge ( conflict / Need to be updated etc.. )
			// if shouldMerge {
			// 	_, _, err := client.PullRequests.Merge(
			// 		context.Background(),
			// 		owner,
			// 		repo,
			// 		*request.Number,
			// 		"chore(deps): Update dep from Dependabot",
			// 		&github.PullRequestOptions{},
			// 	)
			// 	if err != nil {
			// 		fmt.Println("Enable to Merge pull request")
			// 		return err
			// 	}

			// 	PR_MERGED += 1
			// }
		} else {
			PR_NEEDED_ATTENTION = append(
				PR_NEEDED_ATTENTION,
				pull.GetHTMLURL(),
			)
		}
	}

	return nil
}

// Checl if we can approve the PR
func isPullsApprovable(client *github.Client, owner, repo, user string, pullNumber int) (bool, error) {
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
		// if user already approved the PR
		if review.GetState() == APPROVED && review.User.GetLogin() == user {
			return false, nil
		}
	}

	return true, nil
}

// Parse all workflows for this branch
// TODO: Make sure to take laste
func isWorkflowSuccessfull(client *github.Client, owner, repo, branchName string) (bool, error) {
	workflows, _, err := client.Actions.ListRepositoryWorkflowRuns(
		context.TODO(),
		owner,
		repo,
		&github.ListWorkflowRunsOptions{
			Branch: branchName,
		},
	)
	if err != nil {
		return false, err
	}

	onError := 0
	for _, workflow := range workflows.WorkflowRuns {
		if *workflow.Conclusion != "success" {
			onError += 1
		}
	}

	return onError == 0, nil
}

// Parsing a Github response and going threw everything to return all data
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
