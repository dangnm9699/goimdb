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
	"strings"
	"time"
)

const (
	BaseURL = "https://www.imdb.com/title/"
)

var (
	StartPointer *int
	LimitPointer *int
	IDs          []string
	Len          int
)

func init() {
	StartPointer = flag.Int("start", 0, "Start IMDB ID")
	LimitPointer = flag.Int("limit", 100000, "Stop IMDB ID")
	f, _ := ioutil.ReadFile("id.txt")
	fStr := string(f)
	IDs = strings.Split(fStr, "\n")
	Len = len(IDs)
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
		if count > Len {
			break
		}
		id := IDs[count]
		link := BaseURL + id + "/fullcredits"
		//link := "https://www.imdb.com/title/tt4154796/fullcredits?"
		var credits model.Credit
		credits.ID = id
		log.Println(id)
		//credits.ID = "tt4154796"
		credits.CastCrew = make(map[string]interface{})
		c := colly.NewCollector()
		//
		c.OnResponse(func(r *colly.Response) {
			if r.StatusCode != http.StatusOK {
				logger.WriteLog(fmt.Sprintln("[DEBUG] Status Code", r.StatusCode))
			}
		})
		//
		c.OnHTML("div.article.listo", func(e *colly.HTMLElement) {
			// Get general info
			// ... Get poster
			credits.Poster = e.ChildAttr("div.subpage_title_block > a > img", "src")
			// ... Get name
			credits.Name = e.ChildText("div.subpage_title_block > div.subpage_title_block__right-column > div.parent > h3 > a")
			// ... Get year
			credits.Year = model.ExtractYear(e.ChildText("div.subpage_title_block > div.subpage_title_block__right-column > div.parent > h3 > span"))
			// Get full cast & crew info
			keys := make([]string, 0)
			e.ForEach("div.header > h4", func(i int, el *colly.HTMLElement) {
				k := el.Attr("name")
				keys = append(keys, k)
			})
			e.ForEach("div.header > table", func(i int, el *colly.HTMLElement) {
				k := keys[i]
				if k == "director" {
					people := make([]model.Director, 0)
					el.ForEach("tbody > tr", func(i int, elm *colly.HTMLElement) {
						var p model.Director
						p.ID = model.GetPersonID(elm.ChildAttr("td > a", "href"))
						p.Name = strings.TrimSpace(elm.ChildText("td > a"))
						people = append(people, p)
					})
					credits.CastCrew[k] = people
				} else if k == "cast" {
					people := make([]model.Cast, 0)
					el.ForEach("tbody > tr", func(i int, elm *colly.HTMLElement) {
						cl := elm.Attr("class")
						if cl == "odd" || cl == "even" {
							var p model.Cast
							elm.ForEach("td", func(i int, elem *colly.HTMLElement) {
								cla := elem.Attr("class")
								if cla == "primary_photo" {
									p.ID = model.GetPersonID(elem.ChildAttr("a", "href"))
									p.Name = strings.TrimSpace(elem.ChildAttr("a > img", "title"))
									p.Photo = elem.ChildAttr("a > img", "src")
								}
								if cla == "character" {
									characters := strings.TrimSpace(elem.Text)
									characters = strings.ReplaceAll(characters, "\n", "")
									characters = strings.ReplaceAll(characters, "\t", "")
									characters = strings.ReplaceAll(characters, "              ", " ")
									characters = strings.ReplaceAll(characters, "       ", " ")
									//log.Println(characters)
									p.Character = characters
								}
							})
							people = append(people, p)
						}
					})
					credits.CastCrew[k] = people
				} else {
					people := make([]model.Person, 0)
					el.ForEach("tbody > tr", func(i int, elm *colly.HTMLElement) {
						var p model.Person
						elm.ForEach("td", func(i int, elem *colly.HTMLElement) {
							cla := elem.Attr("class")
							if cla == "name" {
								p.ID = model.GetPersonID(elem.ChildAttr("a", "href"))
								p.Name = strings.TrimSpace(elem.ChildText("a"))
							}
							if cla == "credit" {
								p.Credit = strings.TrimSpace(elem.Text)
							}
						})
						people = append(people, p)
					})
					credits.CastCrew[k] = people
				}

			})
			log.Println(credits)
			db.ReplaceOne(credits)
		})
		//
		_ = c.Visit(link)
	}
	log.Println("[INFO] Done, wait for 10 seconds")
	time.Sleep(10 * time.Second)
}
