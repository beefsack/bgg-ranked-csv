package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/beefsack/go-geekdo"
)

func main() {
	stderr := log.New(os.Stderr, "", 0)
	client := geekdo.NewClient()
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
		stderr.Fatalf("Error writing line of CSV, %v", err)
	}
	page := 1
	fetching := true
	for fetching {
		url := fmt.Sprintf(
			"http://boardgamegeek.com/search/boardgame/page/%d?sort=rank&advsearch=1&nosubtypes%%5B0%%5D=boardgameexpansion",
			page,
		)
		r, err := client.AdvSearch(url)
		if err != nil {
			stderr.Fatalf("Error querying %s, %v", url, err)
		}
		if len(r) == 0 {
			break
		}
		for _, thing := range r {
			if thing.Rank == 0 {
				fetching = false
				break
			}
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
				stderr.Fatalf("Error writing line of CSV, %v", err)
			}
		}
		// Rate limit
		time.Sleep(5 * time.Second)
	}
}
