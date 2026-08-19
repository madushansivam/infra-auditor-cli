package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

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

func main() {
	target := flag.String("target", "", "owner/repo to audit")
	flag.Parse()

	if *target == "" {
		fmt.Println("Error: --target is required (format: owner/repo)")
		os.Exit(1)
	}

	owner, name, found := strings.Cut(*target, "/")
	if !found {
		fmt.Println("Error: --target must be in the format owner/repo")
		os.Exit(1)
	}

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
	ctx := context.Background()

	result, err := auditRepo(ctx, client, owner, name)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Repo: %s\nOpen issues: %d\n", result.Name, result.OpenIssues)
}
