package main

import (
	"flag"
	"fmt"
	"github.com/gocolly/colly"
	"imdb/db"
	"imdb/logger"
	"imdb/model"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	BaseURL        = "https://www.imdb.com/title/"
	StopThreshHold = 10
)

var StartPointer *int

func init() {
	StartPointer = flag.Int("start", 1, "Start IMDB ID")
}

func main() {
	defer func() {
		db.DisconnectDB()
		_ = logger.F.Close()
	}()
	fmt.Println(db.ListDatabases())
	flag.Parse()
	start := *StartPointer
	for count := start; true; count++ {
		log.Println("Movie", count)
		id := genId(count)
		link := BaseURL + "tt" + id + "/"
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
			movie.Name = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.titleBar > div.title_wrapper > h1")
			// get Rating
			movie.Rating = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.ratings_wrapper > div.imdbRating > div.ratingValue > strong > span[itemprop=ratingValue]")
			// get RatingCount
			movie.RatingCount = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.ratings_wrapper > div.imdbRating > a > span[itemprop=ratingCount]")
			// get List of Genres
			var genres []string
			e.ForEach("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.titleBar > div.title_wrapper > div.subtext > a[href]", func(i int, element *colly.HTMLElement) {
				genres = append(genres, element.Text)
			})
			if len(genres) > 0 {
				movie.Genres = genres[:len(genres)-1]
			}
			// get Poster
			p1 := e.ChildAttr("div#main_top > div.title-overview > div#title-overview-widget > div.posterWithPlotSummary > div.poster > a > img", "src")
			p2 := e.ChildAttr("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.slate_wrapper > div.poster > a > img", "src")
			if p1 != "" {
				movie.Poster = p1
			}
			if p2 != "" {
				movie.Poster = p2
			}
			// get Budget
			budget := e.ChildText("div#main_bottom > div#titleDetails > div.txt-block:contains('Budget:')")
			movie.Budget = getMoney(budget)
			// get Cumulative
			cumulative := e.ChildText("div#main_bottom > div#titleDetails > div.txt-block:contains('Cumulative Worldwide Gross:')")
			movie.Cumulative = getMoney(cumulative)
			// get Director
			movie.Director = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.plot_summary_wrapper > div.plot_summary > div.credit_summary_item > h4.inline:contains('Director:') ~ a[href]")
			// get Stars
			var stars []string
			e.ForEach("div#main_top > div.title-overview > div#title-overview-widget > div.plot_summary_wrapper > div.plot_summary > div.credit_summary_item > h4.inline:contains('Stars:') ~ a[href]", func(i int, element *colly.HTMLElement) {
				if element.Text != "See full cast & crew" {
					stars = append(stars, element.Text)
				}
			})
			movie.Stars = stars
			// get Country
			country := e.ChildText("div#main_bottom > div#titleDetails > div.txt-block > h4.inline:contains('Country:') ~ a[href]")
			movie.Country = country
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

func genId(stt int) string {
	a := strconv.Itoa(stt)
	if len(a) > 7 {
		return a
	}
	return strings.Repeat("0", 7-len(a)) + a
}
