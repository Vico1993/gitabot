package main

import (
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

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN not found, update your .env")
		log.Fatalln(err)
	}

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
	for _, cRepo := range config.Repos {
		WAIT_GROUP.Add(1)
		cRepo := cRepo

		repository := InitRepository(
			client,
			cRepo.Owner,
			cRepo.Repo,
			config.User,
			cRepo.Merge,
		)

		go func() {
			defer WAIT_GROUP.Done()

			err := repository.HandleDependabotPulls()
			if err != nil {
				log.Fatalln(err)
			}

			if !repository.shouldMerge && repository.pullToMerge != 0 {
				fmt.Println("Done: ", cRepo.Owner, cRepo.Repo)
				return
			}

			err = repository.HandleMerge()
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println("Done: ", cRepo.Owner, cRepo.Repo)
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
