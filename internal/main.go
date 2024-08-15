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

	repositories, err := initConfig()
	if err != nil {
		fmt.Println("Issue retriving config")
		log.Fatalln(err)
	}
	service.Telegram.PostMessage("🤖 🚧 [Gitabot]: Starting...")

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
	for _, repo := range repositories {
		WAIT_GROUP.Add(1)
		repo := repo

		go func() {
			defer WAIT_GROUP.Done()

			err := fetchPullRequests(client, repo.Owner, repo.Repo)
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

	service.Telegram.PostMessage("🤖 🟩 [Gitabot]: Number of PR approved " + strconv.Itoa(PR_APPROVED) + " \n 🤖 🟪 [Gitabot]: Number of PR merged " + strconv.Itoa(PR_MERGED))

	service.Telegram.PostMessage("🤖 ✅ [Gitabot]: Done")
	fmt.Println("Done!")
}

// Fetch pull request from Repository and approving the repo
func fetchPullRequests(client *github.Client, owner, repo string) error {
	requests, _, err := client.PullRequests.List(
		context.Background(),
		owner,
		repo,
		&github.PullRequestListOptions{
			State: "open",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		})

	if err != nil {
		fmt.Println("Enable to fetch pull requests")
		return err
	}

	// TODO: Implement logic to fetch all pages if needed
	// fmt.Println("Last Page: ", response.LastPage)

	for _, request := range requests {
		// Filter none dependabot PR
		if *request.User.Type != DEPENDABOT_TYPE || *request.User.Login != DEPENDABOT_LOGIN || *request.Draft {
			continue
		}

		isMergeable, err := isWorkflowSuccessfull(client, owner, repo, *request.Head.Ref)
		if err != nil {
			fmt.Println("Enable to fetch pull request workflows")
			return err
		}

		if isMergeable {
			// TODO: If already approved no need to re-approved
			_, _, err := client.PullRequests.CreateReview(
				context.TODO(),
				owner,
				repo,
				*request.Number,
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
			PR_NEEDED_ATTENTION = append(PR_NEEDED_ATTENTION, *request.URL)
		}
	}

	return nil
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
