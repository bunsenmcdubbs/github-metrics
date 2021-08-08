package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v37/github"
	_ "github.com/mattn/go-sqlite3"
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

func readArgs() (dbpath, owner, repo string) {
	if len(os.Args) != 4 {
		panic("incorrect number of arguments. usage: list-open-prs <sqlite.db> <owner> <repo>")
	}
	return os.Args[1], os.Args[2], os.Args[3]
}

func createDB(ctx context.Context, path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery := `
CREATE TABLE prs (
    repository text not null,
	id integer not null,
	author text not null,
	title text,
	created_at integer not null,
	closed_at integer,
	
	primary key (repository, id)
)
`

	_, err = db.ExecContext(ctx, createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func insertPR(ctx context.Context, db *sql.DB, repo string, pr *github.PullRequest) error {
	query := `
INSERT INTO prs (repository, id, author, title, created_at, closed_at)
VALUES(?, ?, ?, ?, ?, ?);
`

	closedAt := sql.NullString{
		String: pr.GetClosedAt().Format(time.RFC3339),
		Valid:  !pr.GetClosedAt().IsZero(),
	}
	_, err := db.ExecContext(ctx, query,
		repo,
		pr.GetNumber(),
		pr.GetUser().GetLogin(),
		pr.GetTitle(),
		pr.GetCreatedAt().Format(time.RFC3339),
		closedAt)

	return err
}

func main() {
	ctx := context.Background()
	httpClient := authedHTTPClient(ctx, readGithubToken())
	client := github.NewClient(httpClient)

	dbpath, owner, repo := readArgs()

	db := createDB(ctx, dbpath)
	defer db.Close()

	opts := &github.PullRequestListOptions{
		State:       "closed",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		prs, resp, err := client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to list PR's: %w", err))
		}

		for _, pr := range prs {
			if err := insertPR(ctx, db, owner+"/"+repo, pr); err != nil {
				log.Fatal("failed to insert PR into database", err)
			}
		}

		log.Println("inserted records. count:", len(prs))

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
}
