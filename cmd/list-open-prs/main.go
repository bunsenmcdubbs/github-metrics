package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/oauth2"
)

func readGithubToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func authedHTTPClient(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, ts)
}

func readArgs() (owner string, repo string) {
	if len(os.Args) != 3 {
		fmt.Println(os.Args)
		panic("incorrect number of arguments. usage: list-open-prs <owner> <repo>")
	}
	return os.Args[1], os.Args[2]
}

func main() {
	ctx := context.Background()
	httpClient := authedHTTPClient(ctx, readGithubToken())
	client := github.NewClient(httpClient)

	owner, repo := readArgs()

	prs, _, err := client.PullRequests.List(ctx, owner, repo, nil)
	if err != nil {
		panic(fmt.Errorf("failed to list PR's: %w", err))
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Author", "Title", "Created At", "Open Time"})

	for _, pr := range prs {
		table.Append([]string{
			strconv.Itoa(int(pr.GetNumber())),
			pr.GetUser().GetLogin(),
			pr.GetTitle(),
			pr.GetCreatedAt().In(time.Local).Format(time.RFC1123),
			time.Now().Sub(pr.GetCreatedAt()).Round(time.Second).String(),
		})
	}

	table.Render()
}
