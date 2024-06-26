package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/beefsack/go-geekdo"
)

const MAX_RETRIES = 5

func main() {
	stderr := log.New(os.Stderr, "", log.Ldate|log.Ltime)
	client, err := geekdo.NewClient()
	if err != nil {
		stderr.Fatalf("Error creating client, %v\n", err)
	}
	username := os.Getenv("BGG_USERNAME")
	password := os.Getenv("BGG_PASSWORD")
	if username != "" && password != "" {
		stderr.Println("Logging in")
		if err = client.Login(username, password); err != nil {
			stderr.Fatalf("Error logging in, %v\n", err)
		}
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()
	if err := w.Write([]string{
		"ID",
		"Name",
		"Year",
		"Rank",
		"Average",
		"Bayes average",
		"Users rated",
		"URL",
		"Thumbnail",
	}); err != nil {
		stderr.Fatalf("Error writing line of CSV, %v\n", err)
	}
	page := 0
	for {
		page++
		url := fmt.Sprintf(
			"https://boardgamegeek.com/browse/boardgame/page/%d?sort=rank&sortdir=asc",
			page,
		)
		stderr.Printf("GET %s\n", url)
		r := []geekdo.SearchCollectionItem{}
		tries := 0
		for {
			tries++
			// The request can sometimes fail, or hit a cache with an empty page, so
			// we will have a number of retries
			var err error
			r, err = client.AdvSearch(url)
			if err != nil {
				stderr.Printf("Error querying %s, %v\n", url, err)
				if tries == MAX_RETRIES {
					os.Exit(1)
				}
				continue
			}
			if len(r) == 0 {
				stderr.Println("No results")
				// We will still retry because BGG sometimes has a caching issue
				// where no ranked games appear on an index page
				if tries == MAX_RETRIES {
					break
				}
				continue
			}
			// Check for any unranked games, which will also trigger a retry
			for _, thing := range r {
				if thing.Rank == 0 {
					stderr.Println("Result contains some unranked games")
					// We will still retry because BGG sometimes has a caching issue
					// where no ranked games appear on an index page
					if tries == MAX_RETRIES {
						break
					}
					continue
				}
			}
			// We got here with a full page of ranked games
			break
		}
		rankedOnPage := 0
		for _, thing := range r {
			if thing.Rank == 0 {
				continue
			}
			rankedOnPage++
			if err := w.Write([]string{
				fmt.Sprintf("%v", thing.ID),
				thing.Name,
				fmt.Sprintf("%v", thing.Year),
				fmt.Sprintf("%v", thing.Rank),
				fmt.Sprintf("%v", thing.Average),
				fmt.Sprintf("%v", thing.BayesAverage),
				fmt.Sprintf("%v", thing.UsersRated),
				thing.URL,
				thing.Thumbnail,
			}); err != nil {
				stderr.Fatalf("Error writing line of CSV, %v\n", err)
			}
		}
		if rankedOnPage == 0 {
			stderr.Printf("No ranked games on page %d, stopping\n", page)
			break
		}
	}
	stderr.Printf("Finished on page %d\n", page)
}
