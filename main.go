package main

import (
	"context"
	"flag"
	humanize "github.com/dustin/go-humanize"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"html/template"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type RepoInfo struct {
	Repository  *github.Repository
	URL         string
	IssueCount  int
	Name        string
	Description string
	LastUpdated string
}

func now() string {
	return time.Now().Format(time.RFC3339)
}

func main() {
	// Define the UI
	var token, outflag string
	flag.StringVar(&token, "t", "", "GitHub Token")
	flag.StringVar(&outflag, "o", "", "Path to the rendered template")
	flag.Parse()

	var out *os.File
	var err error
	if outflag == "" {
		out = os.Stdout
	} else {
		out, err = os.Create(outflag)
		if err != nil {
			log.Fatalf("Could not open %v: %v", outflag, err)
		}
	}

	// Setup GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opts := &github.SearchOptions{
		Sort:        "updated",
		ListOptions: github.ListOptions{PerPage: 100, Page: 1},
	}

	// Search for issues labelled "Hacktoberfest" across Julia repos
	var issues []*github.Issue
	for {
		res, resp, err := client.Search.Issues(ctx, "is:issue is:open language:julia label:hacktoberfest", opts)
		if err != nil {
			log.Fatalf("Could not execute search: %v", err)
		} else if resp.StatusCode != 200 {
			log.Fatalf("Search failed: %v - %v", resp.StatusCode, resp.Status)
		}

		log.Printf("Searching page %v/%v...", opts.Page, int(math.Ceil(float64(res.GetTotal())/float64(100))))

		issues = append(issues, res.Issues...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	repos := make(map[string]int)

	for _, issue := range issues {
		repos[issue.GetRepositoryURL()]++
	}

	// Create entries for the list of participating repos
	var entries []RepoInfo
	for repoURL, count := range repos {
		// Get repo data from API
		splitURL := strings.Split(repoURL, "/")
		owner := splitURL[len(splitURL)-2]
		repo := splitURL[len(splitURL)-1]

		log.Printf("Fetching repo info for %v/%v...", owner, repo)
		res, resp, err := client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			log.Fatalf("Could not get repo: %v", err)
		} else if resp.StatusCode != 200 {
			log.Fatalf("Repo lookup failed: %v - %v", resp.StatusCode, resp.Status)
		}

		entries = append(entries, RepoInfo{
			Repository:  res,
			URL:         res.GetHTMLURL(),
			IssueCount:  count,
			Name:        res.GetFullName(),
			Description: res.GetDescription(),
			LastUpdated: humanize.Time(res.GetUpdatedAt().Time),
		})
	}

	// Sort entries by time of last update
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Repository.GetUpdatedAt().Time.After(entries[j].Repository.GetUpdatedAt().Time) })

	// Remove repos that haven't had any activity in the last 6 months
	filteredEntries := []RepoInfo{}
	for _, r := range entries {
		if r.Repository.GetUpdatedAt().After(time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC)) {
			filteredEntries = append(filteredEntries, r)
		}
	}

	// Render template
	funcMap := template.FuncMap{"now": now}
	log.Println("Parsing template...")
	tpl, err := template.New("template.html").Funcs(funcMap).ParseFiles("template.html")
	if err != nil {
		log.Fatalf("Could not parse template: %v", err)
	}

	log.Println("Executing template...")
	err = tpl.Execute(out, filteredEntries)
	if err != nil {
		log.Fatalf("Could not execute template: %v", err)
	}
}
