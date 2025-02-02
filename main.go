package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"imdb/db"
	"imdb/logger"
	"imdb/model"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BaseURL = "https://www.imdb.com/title/"
	con     = 10
	req     = 100000
)

var (
	r  *regexp.Regexp
	ch chan model.Movie
)

func init() {
	r, _ = regexp.Compile("(.+)\\((\\d{4})\\)")
	ch = make(chan model.Movie, 1000)
}

func extractName(name string) (string, string) {
	rr := r.FindStringSubmatch(name)
	if len(rr) == 3 {
		return strings.TrimSpace(rr[1]), rr[2]
	}
	return strings.TrimSpace(name), ""
}

func main() {
	defer func() {
		db.DisconnectDB()
		_ = logger.F.Close()
	}()
	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if ok {
					db.ReplaceOne(msg)
				}
			}
		}
	}()
	fmt.Println("[INFO] List DB:", db.ListDatabases())
	var wg sync.WaitGroup
	wg.Add(con)
	for i := 0; i < con; i++ {
		go func(idx int) {
			for count := idx * req; count < (idx+1)*req; count++ {
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
					name := e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.titleBar > div.title_wrapper > h1")
					movie.Name, movie.Year = extractName(name)
					if movie.Year == "" {
						// item is not a movie, continue
						return
					}
					// get Rating
					movie.Rating = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.ratings_wrapper > div.imdbRating > div.ratingValue > strong > span[itemprop=ratingValue]")
					// get RatingCount
					movie.RatingCount = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.ratings_wrapper > div.imdbRating > a > span[itemprop=ratingCount]")
					// get List of Genres
					var genres string
					e.ForEach("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.titleBar > div.title_wrapper > div.subtext > a", func(i int, element *colly.HTMLElement) {
						if element.Attr("title") == "" {
							genres = addToString(genres, element.Text)
						}
					})
					movie.Genres = genres
					// get Duration
					duration := e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.title_block > div.title_bar_wrapper > div.titleBar > div.title_wrapper > div.subtext > time")
					movie.Duration = strings.TrimSpace(duration)
					// get Poster
					//p1 := e.ChildAttr("div#main_top > div.title-overview > div#title-overview-widget > div.posterWithPlotSummary > div.poster > a > img", "src")
					//p2 := e.ChildAttr("div#main_top > div.title-overview > div#title-overview-widget > div.vital > div.slate_wrapper > div.poster > a > img", "src")
					//if p1 != "" {
					//	movie.Poster = p1
					//}
					//if p2 != "" {
					//	movie.Poster = p2
					//}
					// get Budget
					budget := e.ChildText("div#main_bottom > div#titleDetails > div.txt-block:contains('Budget:')")
					movie.Budget = getMoney(budget)
					// get Cumulative
					cumulative := e.ChildText("div#main_bottom > div#titleDetails > div.txt-block:contains('Cumulative Worldwide Gross:')")
					movie.Cumulative = getMoney(cumulative)
					// get Director
					var director string
					director = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.plot_summary_wrapper > div.plot_summary > div.credit_summary_item > h4.inline:contains('Director:') ~ a[href]")
					if len(director) == 0 {
						director = e.ChildText("div#main_top > div.title-overview > div#title-overview-widget > div.posterWithPlotSummary > div.plot_summary_wrapper > div.plot_summary > div.credit_summary_item > h4.inline:contains('Director:') ~ a[href]")
					}
					movie.Director = director
					// get Stars
					var stars string
					e.ForEach("div#main_top > div.title-overview > div#title-overview-widget > div.plot_summary_wrapper > div.plot_summary > div.credit_summary_item > h4.inline:contains('Star') ~ a[href]", func(i int, element *colly.HTMLElement) {
						if element.Text != "See full cast & crew" {
							stars = addToString(stars, element.Text)
						}
					})
					if stars == "" {
						e.ForEach("div#main_top > div.title-overview > div#title-overview-widget > div.posterWithPlotSummary > div.plot_summary_wrapper > div.plot_summary > div.credit_summary_item > h4.inline:contains('Star') ~ a[href]", func(i int, element *colly.HTMLElement) {
							if element.Text != "See full cast & crew" {
								stars = addToString(stars, element.Text)
							}
						})
					}
					movie.Stars = stars
					// get Country
					var country string
					e.ForEach("div#main_bottom > div#titleDetails > div.txt-block > h4.inline:contains('Country:') ~ a[href]", func(i int, element *colly.HTMLElement) {
						country = addToString(country, strings.TrimSpace(element.Text))
					})
					movie.Country = country
					// get StoryLine
					storyLine := e.ChildText("div#main_bottom > div#titleStoryLine > div.inline.canwrap > p > span")
					movie.StoryLine = strings.TrimSpace(storyLine)
					ch <- movie
				})
				//
				_ = c.Visit(link)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	log.Println("[INFO] Done, wait for 10 seconds")
	time.Sleep(10 * time.Second)
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

func addToString(dst, app string) string {
	if dst != "" {
		dst = dst + "," + app
	} else {
		dst = app
	}
	return dst
}
