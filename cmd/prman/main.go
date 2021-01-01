package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"

	"github.com/fatih/color"
	"github.com/shric/prman/pkg/client"
)

type Pull struct {
	pr       *github.PullRequest
	reviews  []*github.PullRequestReview
	statuses []*github.RepoStatus
	apicount int
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <github search>", os.Args[0])
	}
	token := os.Getenv("GITHUB_API_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	c := client.NewClientWithContext(ctx, tc)

	issues := c.SearchIssues("is:pr " + os.Args[1])
	apicount := 1
	fmt.Println("Found", len(issues), "pull requests")

	ch := make(chan *Pull)
	var wg sync.WaitGroup
	for _, issue := range issues {
		wg.Add(1)
		go func(issue github.Issue, ch chan *Pull) {
			defer wg.Done()
			result := &Pull{}

			pr, err, count := c.GetPullRequestFromIssue(&issue)
			result.apicount += count
			if err != nil {
				panic(err)
			}
			result.pr = pr

			reviews, err, count := c.GetReviewsFromPR(pr)
			result.apicount += count
			if err != nil {
				panic(err)
			}
			result.reviews = reviews

			statuses, err, count := c.GetStatusesFromPR(pr)
			result.apicount += count
			if err != nil {
				panic(err)
			}
			result.statuses = statuses
			ch <- result
		}(*issue, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []*Pull
	for result := range ch {
		results = append(results, result)
		apicount += result.apicount
	}
	fmt.Println("Made", apicount, "API calls to github.com.")

	sort.Slice(results, func(a, b int) bool {
		return *results[a].pr.URL < *results[b].pr.URL
	})

	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	hiGreen := color.New(color.FgHiGreen).SprintFunc()
	hiBlue := color.New(color.FgHiBlue).SprintFunc()

	for _, result := range results {
		var state string
		pr := result.pr
		if *pr.State == "open" {
			switch *pr.MergeableState {
			case "blocked":
				state = hiBlue("blocked")
			case "clean":
				state = green("ready")
			case "draft":
				state = yellow("draft")
			case "unstable":
				state = red("unstable")
			default:
				state = red(*pr.MergeableState)
			}
		} else {
			if *pr.Merged {
				state = hiGreen("merged")
			} else {
				state = red("closed")
			}

		}

		var approveStatus string
		approver, ok := find_approver(result.reviews); if ok {
			approveStatus = green("approved by "+ approver)
		} else {
			approveStatus = yellow("needs approval")
		}

		fmt.Printf("%-20s%-40s%s\n", state, approveStatus, *pr.HTMLURL)
		for _, status := range result.statuses {
			if *status.State != "pending" && *status.State != "success" {
				fmt.Printf("    %s\n        %s\n", red(*status.Context + ": " + *status.Description), *status.TargetURL)
			}

		}
		/*
		for _, review := range result.reviews {
			switch *review.State {
			case "CHANGES_REQUESTED":
				state = yellow("CHANGES REQUESTED")
			case "APPROVED":
				state = green("APPROVED")
			case "COMMENTED":
				state = hiBlue("COMMENTED")
			default:
				state = *review.State
			}
			fmt.Printf("    %-15s %-28s %s\n", *review.User.Login, state, *review.Body)

		}
		 */
	}

}

func find_approver(reviews []*github.PullRequestReview) (string, bool) {
	for _, review := range reviews {
		if *review.State == "APPROVED" {
			return *review.User.Login, true
		}
	}
	return "", false
}
