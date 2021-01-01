package client

import (
	"context"
	"errors"
	"github.com/google/go-github/v33/github"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	ctx    context.Context
	client *github.Client
}

func NewClientWithContext(ctx context.Context, c *http.Client) *Client {
	client := github.NewClient(c)

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

func parsePRUrl(issue *github.Issue) (owner string, repo string, number int, err error) {
	if !issue.IsPullRequest() {
		err = errors.New("not a PR")
		return
	}
	split := strings.Split(*issue.PullRequestLinks.URL, "/")
	owner = split[4]
	repo = split[5]
	number, err = strconv.Atoi(split[7])
	return
}

func (c *Client) GetPullRequestFromIssue(issue *github.Issue) (*github.PullRequest, error, int) {
	owner, repo, number, err := parsePRUrl(issue)
	if err != nil {
		return nil, err, 0
	}
	pr, _, err := c.client.PullRequests.Get(c.ctx, owner, repo, number)
	if err != nil {
		return nil, err, 1
	}
	return pr, nil, 1
}

func (c *Client) ListReviewComments(issue *github.Issue, id int64) (result []*github.PullRequestComment, err error, count int) {
	opts := &github.ListOptions{PerPage: 100}

	for {
		count++
		prc, resp, err2 := c.client.PullRequests.ListReviewComments(c.ctx, *issue.Repository.Owner.Login, *issue.Repository.Name, *issue.Number, id, opts)
		if err2 != nil {
			err = err2
			return
		}
		result = append(result, prc...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return
}

func (c *Client) GetStatusesFromPR(pr *github.PullRequest) (result []*github.RepoStatus, err error, count int) {
	if pr.StatusesURL == nil {
		err = errors.New("no status")
		return
	}
	opts := &github.ListOptions{PerPage: 100}
	for {
		rs, resp, err2 := c.client.Repositories.ListStatuses(c.ctx, *pr.Base.Repo.Owner.Login, *pr.Base.Repo.Name, *pr.Head.SHA, opts)
		if err2 != nil {
			err = err2
			return
		}
		result = append(result, rs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return
}

func (c *Client) GetReviewsFromPR(pr *github.PullRequest) (result []*github.PullRequestReview, err error, count int) {
	opts := &github.ListOptions{PerPage: 100}
	for {
		count++
		reviews, resp, err2 := c.client.PullRequests.ListReviews(c.ctx, *pr.Base.Repo.Owner.Login, *pr.Base.Repo.Name, *pr.Number, opts)
		if err2 != nil {
			err = err2
			return
		}
		result = append(result, reviews...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return
}

func (c *Client) SearchIssues(query string) []*github.Issue {
	opts := &github.SearchOptions{
		Sort:        "created",
		Order:       "asc",
		TextMatch:   false,
		ListOptions: github.ListOptions{PerPage:100},
	}
	var allIssues []*github.Issue
	for {
		issues, resp, err := c.client.Search.Issues(c.ctx, query, opts)
		if err != nil {
			panic(err)
		}
		allIssues = append(allIssues, issues.Issues...)
		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage

	}
	return allIssues
}

func (c *Client) GetAllReposByOrg(org string) []*github.Repository {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository

	for {
		repos, resp, err := c.client.Repositories.ListByOrg(c.ctx, org, opt)

		if err != nil {
			panic(err)
		}
		for _, repo := range repos {
			if !*repo.Archived {
				allRepos = append(allRepos, repo)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos
}
