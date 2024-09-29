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

	fmt.Println("🤖 🚧: Starting...")

	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")
	if token == "" || username == "" {
		fmt.Println("GITHUB_TOKEN or GITHUB_USERNAME not found, update your .env")
		log.Fatalln(err)
	}

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
	for _, cRepo := range d {
		WAIT_GROUP.Add(1)
		cRepo := cRepo

		fmt.Println("Start looking at: ", cRepo.owner, cRepo.name)

		repository := InitRepository(
			client,
			cRepo.owner,
			cRepo.name,
			username,
			cRepo.allowToMerge,
		)

		go func() {
			defer WAIT_GROUP.Done()

			err := repository.HandleDependabotPulls()
			if err != nil {
				log.Fatalln(err)
			}

			if !repository.shouldMerge || repository.pullToMerge == 0 {
				fmt.Println("Done: ", cRepo.owner, cRepo.name)
				return
			}

			err = repository.HandleMerge()
			if err != nil {
				fmt.Println("Error Merging: ", cRepo.owner, cRepo.name)
				fmt.Println(err)
			}

			fmt.Println("Done: ", cRepo.owner, cRepo.name)
		}()
	}

	WAIT_GROUP.Wait()

	txt := "🟩 Number of PR approved " + strconv.Itoa(PR_APPROVED) + " \n🟪 Number of PR merged " + strconv.Itoa(PR_MERGED)

	if PR_MERGED > 0 {
		fmt.Println("Pull Requests merged: ", PR_MERGED)
		fmt.Println("Pull Requests merged - Error: ", PR_MERGED_ERROR)
	}

	if len(PR_NEEDED_ATTENTION) > 0 {
		txt += "🟥 Number of PR that need attention " + strconv.Itoa(len(PR_NEEDED_ATTENTION))
		fmt.Println("Few pull requests need your attention")
		fmt.Println(utils.ToJson(PR_NEEDED_ATTENTION))
	}

	logging(txt)

	fmt.Println("Done!")
}

// Log information in telegram
// but also in terminal
func logging(txt string) {
	fmt.Println("{🤖} " + txt)
	err := service.Telegram.PostMessage("🤖 {Gitabot} : \n\n" + txt)
	if err != nil {
		fmt.Println("❌ - Couldn't post in telegram: " + err.Error())
	}
}
