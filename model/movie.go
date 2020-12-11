package model

import (
	"regexp"
	"strings"
)

//type Movie struct {
//	ID          string   `json:"id" bson:"id"`
//	Name        string   `json:"name" bson:"name"`
//	Year        string   `json:"year" bson:"year"`
//	Rating      string   `json:"rating" bson:"rating"`
//	RatingCount string   `json:"rating_count" bson:"rating_count"`
//	Duration    string   `json:"duration" bson:"duration"`
//	Genres      []string `json:"genres" bson:"genres"`
//	Poster      string   `json:"poster" bson:"poster"`
//	Budget      string   `json:"budget" bson:"budget"`
//	Cumulative  string   `json:"cumulative" bson:"cumulative"`
//	Director    string   `json:"director" bson:"director"`
//	Stars       []string `json:"stars" bson:"stars"`
//	Country     []string `json:"country" bson:"country"`
//	StoryLine   string   `json:"story_line" bson:"story_line"`
//}

type Credit struct {
	ID       string                 `json:"tconst" bson:"tconst"`
	Name     string                 `json:"name" bson:"name"`
	Year     string                 `json:"release_year" bson:"release_year"`
	Poster   string                 `json:"poster" bson:"poster"`
	CastCrew map[string]interface{} `json:"cast_and_crew" bson:"cast_and_crew"`
}

type Director struct {
	ID   string `json:"nconst" bson:"nconst"`
	Name string `json:"name" bson:"name"`
}

type Person struct {
	ID     string `json:"nconst" bson:"nconst"`
	Name   string `json:"name" bson:"name"`
	Credit string `json:"credit" bson:"credit"`
}

type Cast struct {
	ID        string `json:"nconst" bson:"nconst"`
	Name      string `json:"name" bson:"name"`
	Photo     string `json:"photo" bson:"photo"`
	Character string `json:"character" bson:"character"`
}

func GetPersonID(a string) string {
	// a Example: /name/nm1321655/?ref_=ttfc_fc_wr1
	splits := strings.Split(a, "/")
	return splits[2]
}

func ExtractYear(year string) string {
	// name Example: Avengers: Endgame (2019)
	r, _ := regexp.Compile("\\((\\d{4})\\)")
	rr := r.FindStringSubmatch(year)
	//if len(rr) == 3 {
	//	return strings.TrimSpace(rr[1]), rr[2]
	//}
	if len(rr) < 2 {
		return ""
	}
	return rr[1]
}
