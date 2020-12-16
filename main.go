package main

import (
	"flag"
	"fmt"
	"github.com/gocolly/colly"
	"imdb/db"
	"imdb/logger"
	"imdb/model"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const (
	BaseURL = "https://www.imdb.com/title/"
)

var (
	StartPointer *int
	LimitPointer *int
	r            *regexp.Regexp
	NFilm        int
	IDs          []string
)

func init() {
	StartPointer = flag.Int("start", 0, "Start IMDB ID")
	LimitPointer = flag.Int("limit", 50000, "Stop IMDB ID")
	r, _ = regexp.Compile("(.+)\\((\\d{4})\\)")
	f, _ := ioutil.ReadFile("id.txt")
	fStr := string(f)
	IDs = strings.Split(fStr, "\n")
	NFilm = len(IDs)
	log.Println("[INFO] Number of Film:", NFilm)
}

func extractName(name string) (string, string) {
	rr := r.FindStringSubmatch(name)
	if len(rr) < 3 {
		return strings.TrimSpace(name), ""
	}
	return strings.TrimSpace(rr[1]), rr[2]
}

func main() {
	defer func() {
		db.DisconnectDB()
		_ = logger.F.Close()
	}()
	fmt.Println("[INFO] List DB:", db.ListDatabases())
	flag.Parse()
	start := *StartPointer
	end := start + *LimitPointer
	for count := start; count < end; count++ {
		if count == NFilm {
			break
		}
		id := IDs[count]
		link := BaseURL + id + "/"
		var movie model.Movie
		movie.ID = id
		c := colly.NewCollector()
		//
		c.OnResponse(func(r *colly.Response) {
			if r.StatusCode != http.StatusOK {
				logger.WriteLog(fmt.Sprintln("[DEBUG] Status Code", r.StatusCode))
			}
		})
		//
		c.OnHTML("div#content-2-wide", func(e *colly.HTMLElement) {
			// get Name
			name := e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.titleBar > div.title_wrapper > h1")
			movie.Name, movie.Year = extractName(name)
			// get Rating
			movie.Rating = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.ratings_wrapper > div.imdbRating > div.ratingValue > strong > span[itemprop=ratingValue]")
			// get RatingCount
			movie.RatingCount = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.ratings_wrapper > div.imdbRating > a > span[itemprop=ratingCount]")
			// get List of Genres
			var genres []string
			e.ForEach("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.titleBar > div.title_wrapper > div.subtext > a", func(i int, element *colly.HTMLElement) {
				if element.Attr("title") == "" {
					genres = append(genres, element.Text)
				}
			})
			movie.Genres = genres
			// get Duration
			duration := e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.titleBar > div.title_wrapper > div.subtext > time")
			movie.Runtime = strings.TrimSpace(duration)
			// get Budget
			budget := e.ChildText("div#main_bottom > div#titleDetails > div.txt-block:contains('Budget:')")
			movie.Budget = getMoney(budget)
			// get Cumulative
			cumulative := e.ChildText("div#main_bottom > div#titleDetails > div.txt-block:contains('Cumulative Worldwide Gross:')")
			movie.Gross = getMoney(cumulative)
			// get Stars
			var stars []string
			e.ForEach("div#main_top > div.title-overview > div#title-overview-widget > div.plot_summary_wrapper > div.plot_summary > div.credit_summary_item > h4.inline:contains('Star') ~ a[href]", func(i int, element *colly.HTMLElement) {
				if element.Text != "See full cast & crew" {
					stars = append(stars, element.Text)
				}
			})
			// get Country
			var country []string
			e.ForEach("div#main_bottom > div#titleDetails > div.txt-block > h4.inline:contains('Country:') ~ a[href]", func(i int, element *colly.HTMLElement) {
				country = append(country, strings.TrimSpace(element.Text))
			})
			movie.Country = country
			// get Language
			var language []string
			e.ForEach("div#main_bottom > div#titleDetails > div.txt-block > h4.inline:contains('Language:') ~ a[href]", func(i int, element *colly.HTMLElement) {
				language = append(language, strings.TrimSpace(element.Text))
			})
			movie.Language = language
			// get StoryLine
			storyLine := e.ChildText("div#main_bottom > div#titleStoryLine > div.inline.canwrap > p > span")
			movie.StoryLine = strings.TrimSpace(storyLine)
			fmt.Println(movie)
			db.ReplaceOne(movie)
		})
		//
		_ = c.Visit(link)
	}
}

func getMoney(p string) string {
	if len(p) == 0 {
		return p
	}
	a := strings.Split(p, ":")
	a = strings.Split(a[1], "(")
	return strings.TrimSpace(a[0])
}
