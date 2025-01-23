package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	service "github.com/Vico1993/gitabot/internal/services"
	"github.com/Vico1993/gitabot/internal/utils"
	"github.com/google/go-github/v63/github"
	"github.com/subosito/gotenv"
)

const PER_PAGE = 100

var (
	WAIT_GROUP          sync.WaitGroup
	PR_APPROVED         = 0
	PR_MERGED           = 0
	PR_NEEDED_ATTENTION = []string{}
	PR_MERGED_ERROR     = 0
	FILTER_LATEST       = "latest"
	DEPENDABOT_LOGIN    = "dependabot[bot]"
	DEPENDABOT_TYPE     = "Bot"
	APPROVE             = "APPROVE"
	APPROVED            = "APPROVED"
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

	fmt.Println("🚧: Starting...")

	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")
	if token == "" || username == "" {
		fmt.Println("GITHUB_TOKEN or GITHUB_USERNAME not found, update your .env")
		log.Fatalln(err)
	}

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))

	fmt.Println("🔎: Fetching for issues...")
	issues, err := findDependabotIssues(client, username)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("✅: Done")

	cows := map[string]*Repository{}
	for _, issue := range issues {
		// Extract Repository information
		// url will look like: "https://github.com/<OWNER>/<repository name>/pull/<number>
		url := issue.GetHTMLURL()

		str := strings.Replace(url, "https://github.com/", "", 1) // <OWNER>/<repository name>/pull/<number>
		tmp := strings.Split(str, "/")

		number, _ := strconv.Atoi(tmp[3])
		// Check if the key exists
		repo, exists := cows[tmp[0]+"/"+tmp[1]]
		if exists {

			repo.AddPulls(number)
		} else {
			cows[tmp[0]+"/"+tmp[1]] = InitRepository(client, tmp[0], tmp[1], username, []int{number})
		}
	}

	fmt.Println("🔎: Found: ", len(issues))
	fmt.Println("🔎: Cross : ", len(cows), " repositories")

	// TODO: Make a Logging function in Repository struct
	for _, repository := range cows {
		WAIT_GROUP.Add(1)
		repository := repository

		fmt.Println("👀 Start looking at: ", repository.owner, repository.name)

		go func() {
			defer WAIT_GROUP.Done()

			err := repository.HandlePulls()
			if err != nil {
				log.Fatalln(err)
			}

			if repository.pullToMerge == 0 {
				fmt.Println("✅: Done ", repository.owner, repository.name)
				return
			}

			err = repository.HandleMerge()
			if err != nil {
				fmt.Println("🔴: Error Merging ", repository.owner, repository.name)
				fmt.Println(err)
			}

			fmt.Println("✅: Done ", repository.owner, repository.name)
		}()
	}

	WAIT_GROUP.Wait()

	txt := "🟩 Number of PR approved " + strconv.Itoa(PR_APPROVED) + " \n🟪 Number of PR merged " + strconv.Itoa(PR_MERGED)

	if PR_MERGED > 0 {
		fmt.Println("Pull Requests merged: ", PR_MERGED)
		fmt.Println("Pull Requests merged - Error: ", PR_MERGED_ERROR)
	}

	if len(PR_NEEDED_ATTENTION) > 0 {
		txt += "\n🟥 Number of PR that need attention " + strconv.Itoa(len(PR_NEEDED_ATTENTION))
		fmt.Println("👀 Few pull requests need your attention")
		fmt.Println(utils.ToJson(PR_NEEDED_ATTENTION))
	}

	logging(txt)

	fmt.Println("✅: Done")
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
