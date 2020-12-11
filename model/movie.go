package model

import (
	"regexp"
	"strings"
)

type Movie struct {
	ID          string   `json:"tconst" bson:"tconst"`
	Name        string   `json:"name" bson:"name"`
	Year        string   `json:"year" bson:"year"`
	Rating      string   `json:"rating" bson:"rating"`
	RatingCount string   `json:"rating_count" bson:"rating_count"`
	StoryLine   string   `json:"story_line" bson:"story_line"`
	Genres      []string `json:"genres" bson:"genres"`
	Country     []string `json:"country" bson:"country"`
	Language    []string `json:"language" bson:"language"`
	Budget      string   `json:"budget" bson:"budget"`
	Gross       string   `json:"gross" bson:"gross"`
	Runtime     string   `json:"runtime" bson:"runtime"`
}

func GetID(a string) string {
	// a Example: /name/nm1321655/?ref_=ttfc_fc_wr1
	splits := strings.Split(a, "/")
	return splits[2]
}

func ExtractName(r *regexp.Regexp, name string) (string, string) {
	rr := r.FindStringSubmatch(name)
	if len(rr) == 3 {
		return strings.TrimSpace(rr[1]), rr[2]
	}
	return strings.TrimSpace(name), ""
}
