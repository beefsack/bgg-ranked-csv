package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/beefsack/go-geekdo"
)

const (
	TstampFormat = "2006-01-02 15:04:05"
)

var (
	Earliest = time.Now().AddDate(0, 0, -90)
	stderr   = log.New(os.Stderr, "", 0)
)

type Item struct {
	Rating       float64 `json:"rating"`
	RatingTstamp string  `json:"rating_tstamp"`
}

type Comments struct {
	Items []Item `json:"items"`
}

func ratings(thing geekdo.SearchCollectionItem) ([]float64, error) {
	page := 1
	ratings := []float64{}
	fetchingPages := true
	for fetchingPages {
		comments, err := ratingPage(thing, page)
		if err != nil {
			return nil, err
		}
		if len(comments.Items) == 0 {
			break
		}
		for _, item := range comments.Items {
			ratedAt, err := time.Parse(TstampFormat, item.RatingTstamp)
			if err != nil {
				return nil, fmt.Errorf("error parsing rating_tstamp '%s', %s", item.RatingTstamp, err)
			}
			if ratedAt.Before(Earliest) {
				fetchingPages = false
				break
			}
			ratings = append(ratings, item.Rating)
		}
		page++
	}
	return ratings, nil
}

func ratingPage(thing geekdo.SearchCollectionItem, page int) (comments Comments, err error) {
	url := fmt.Sprintf("https://boardgamegeek.com/api/collections?ajax=1&comment=1&objectid=%d&objecttype=thing&oneperuser=0&pageid=%d&rated=1&require_review=true&showcount=1000&sort=rating_tstamp", thing.ID, page)
	stderr.Printf("GET %s", url)
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("error fetching ratings for %s (%d) page %d, %v", thing.Name, thing.ID, page, err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&comments); err != nil {
		err = fmt.Errorf("error decoding ratings JSON for %s (%d) page %d, %v", thing.Name, thing.ID, page, err)
	}
	return
}

func main() {
	// Find the 3000 most rated games.
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
		"Recent average",
		"Recent users rated",
	}); err != nil {
		stderr.Fatalf("Error writing line of CSV, %v", err)
	}
	things := []geekdo.SearchCollectionItem{}
	stderr.Print("Finding most rated games")
	for page := 1; page <= 30; page++ {
		url := fmt.Sprintf(
			"http://boardgamegeek.com/search/boardgame/page/%d?sort=numvoters&advsearch=1&sortdir=desc&nosubtypes%%5B0%%5D=boardgameexpansion",
			page,
		)
		stderr.Printf("GET %s", url)
		r, err := client.AdvSearch(url)
		if err != nil {
			stderr.Fatalf("Error querying %s, %v", url, err)
		}
		things = append(things, r...)
	}
	stderr.Print("Got most rated games")

	// Get recent ratings for each games.
	for _, thing := range things {
		stderr.Printf("Getting ratings for %s (%d)", thing.Name, thing.ID)
		rat, err := ratings(thing)
		if err != nil {
			stderr.Fatalf("Error getting ratings for game, %s", err)
		}
		sum := 0.0
		for _, r := range rat {
			sum += r
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
			fmt.Sprintf("%v", sum/float64(len(rat))),
			fmt.Sprintf("%v", len(rat)),
		}); err != nil {
			stderr.Fatalf("Error writing line of CSV, %v", err)
		}
	}
}
