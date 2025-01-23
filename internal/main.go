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
	PR_ALREADY_APPROVED = 0
	PR_MERGED           = 0
	PR_NEEDED_ATTENTION = []string{}
	PR_MERGED_ERROR     = 0
	ERRORS              = []string{}
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

	for _, repository := range cows {
		WAIT_GROUP.Add(1)
		repository := repository

		repository.Logging("👀 Start looking at: ")

		go func() {
			defer WAIT_GROUP.Done()

			err := repository.HandlePulls()
			if err != nil {
				ERRORS = append(ERRORS, err.Error())
				repository.Logging("🔴: Error handling pulls: " + err.Error())
				return
			}

			if repository.pullToMerge == 0 {
				repository.Logging("✅: Done ")
				return
			}

			err = repository.HandleMerge()
			if err != nil {
				repository.Logging("🔴: Error Merging ")
				fmt.Println(err)
				return
			}

			repository.Logging("✅: Done ")
		}()
	}

	WAIT_GROUP.Wait()

	txt := "🟩 Number of PR approved " + strconv.Itoa(PR_APPROVED) + " \n🟩 Number of PR merged " + strconv.Itoa(PR_ALREADY_APPROVED) + " \n🟪 Number of PR merged " + strconv.Itoa(PR_MERGED) + "\n🟥 Number of PR that need attention " + strconv.Itoa(len(PR_NEEDED_ATTENTION))

	fmt.Println(txt)
	if len(PR_NEEDED_ATTENTION) > 0 {
		fmt.Println("⛑️👀 Few pull requests need your attention")
		fmt.Println(utils.ToJson(PR_NEEDED_ATTENTION))
	}

	err = service.Telegram.PostMessage(txt)
	if err != nil {
		fmt.Println("🔴: - Couldn't post in telegram: " + err.Error())
		return
	}

	if len(ERRORS) > 0 {
		fmt.Println("🔴👀 Repositories need your intention")
		fmt.Println(utils.ToJson(ERRORS))

		err = service.Telegram.PostMessage("🔴🔥 Repositories need your intention #" + strconv.Itoa(len(ERRORS)))
		if err != nil {
			fmt.Println("🔴: - Couldn't post in telegram: " + err.Error())
			return
		}
	}

	if PR_MERGED_ERROR == 0 {
		return
	}

	fmt.Println("🔴🔥 Pull Requests merged - Error: ", PR_MERGED_ERROR)
	err = service.Telegram.PostMessage("🔴🔥 Pull Requests merged - Error: " + strconv.Itoa(PR_MERGED_ERROR))

	if err != nil {
		fmt.Println("🔴: - Couldn't post in telegram: " + err.Error())
		return
	}
}
