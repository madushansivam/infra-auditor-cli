package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/v66/github"
)

type AuditResult struct {
	Name       string
	OpenIssues int
}

func auditRepo(ctx context.Context, client *github.Client, owner, name string) (AuditResult, error) {
	repo, resp, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		if resp != nil {
			switch resp.StatusCode {
			case 404:
				return AuditResult{}, fmt.Errorf("repo %s/%s does not exist or you don't have access to it", owner, name)
			case 401:
				return AuditResult{}, fmt.Errorf("authentication failed — check that GITHUB_TOKEN is set and valid")
			case 403:
				return AuditResult{}, fmt.Errorf("forbidden or rate-limited fetching %s/%s — wait a bit and try again", owner, name)
			}
		}
		return AuditResult{}, fmt.Errorf("fetching repo %s/%s: %w", owner, name, err)
	}
	return AuditResult{
		Name:       repo.GetFullName(),
		OpenIssues: repo.GetOpenIssuesCount(),
	}, nil
}

func auditAll(ctx context.Context, client *github.Client, targets []string, workers int) []AuditResult {
	jobs := make(chan string, len(targets))
	results := make(chan AuditResult, len(targets))
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				owner, name, found := strings.Cut(t, "/")
				if !found {
					fmt.Printf("skipping %q: not in owner/repo format\n", t)
					continue
				}
				res, err := auditRepo(ctx, client, owner, name)
				if err != nil {
					fmt.Printf("skipping %s: %v\n", t, err)
					continue
				}
				results <- res
			}
		}()
	}

	for _, t := range targets {
		jobs <- t
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var all []AuditResult
	for r := range results {
		all = append(all, r)
	}
	return all
}

func main() {
	target := flag.String("target", "", "comma-separated owner/repo list to audit")
	flag.Parse()

	if *target == "" {
		fmt.Println("Error: --target is required (format: owner/repo,owner2/repo2,...)")
		os.Exit(1)
	}

	targets := strings.Split(*target, ",")

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
	ctx := context.Background()

	results := auditAll(ctx, client, targets, 5)

	for _, r := range results {
		fmt.Printf("Repo: %s | Open issues: %d\n", r.Name, r.OpenIssues)
	}
}
