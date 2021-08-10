package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/hako/durafmt"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/oauth2"
)

const ageCutoff = 3 * 7 * 24 * time.Hour // 3 weeks

func readGithubToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func authedHTTPClient(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, ts)
}

func readArgs() (owner string) {
	if len(os.Args) != 2 {
		fmt.Println(os.Args)
		panic("incorrect number of arguments. usage: list-open-prs-org <owner>")
	}
	return os.Args[1]
}

type PRWithRepo struct {
	*github.PullRequest
	Repo string
}

func printTable(prs []PRWithRepo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Repo", "ID", "Author", "Title", "Created At", "Open Time"})

	for _, pr := range prs {
		age := time.Now().Sub(pr.GetCreatedAt()).Round(time.Second)

		table.Append([]string{
			pr.Repo,
			strconv.Itoa(int(pr.GetNumber())),
			pr.GetUser().GetLogin(),
			pr.GetTitle(),
			pr.GetCreatedAt().In(time.Local).Format(time.RFC1123),
			durafmt.Parse(age).LimitFirstN(2).String(),
		})
	}

	table.Render()
}

func main() {
	ctx := context.Background()
	httpClient := authedHTTPClient(ctx, readGithubToken())
	client := github.NewClient(httpClient)

	owner := readArgs()

	reposOpts := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var oldPrs []PRWithRepo
	for {
		repos, repoResp, err := client.Repositories.ListByOrg(ctx, owner, reposOpts)
		if err != nil {
			panic(err)
		}

		for _, repo := range repos {
			log.Println("Looking in repo", repo.GetName())
			opts := &github.PullRequestListOptions{
				State:       "open",
				ListOptions: github.ListOptions{PerPage: 100},
			}

			for {
				prs, resp, err := client.PullRequests.List(ctx, owner, repo.GetName(), opts)
				if err != nil {
					panic(fmt.Errorf("failed to list PR's: %w", err))
				}

				for _, pr := range prs {
					age := time.Now().Sub(pr.GetCreatedAt()).Round(time.Second)
					if age > ageCutoff {
						oldPrs = append(oldPrs, PRWithRepo{pr, repo.GetName()})
					}
				}

				if resp.NextPage == 0 {
					break
				}
				opts.Page = resp.NextPage
			}

		}

		if repoResp.NextPage == 0 {
			break
		}
		reposOpts.Page = repoResp.NextPage
	}

	sort.Slice(oldPrs, func(i, j int) bool {
		return oldPrs[i].CreatedAt.Before(*oldPrs[j].CreatedAt)
	})

	printTable(oldPrs)
}
