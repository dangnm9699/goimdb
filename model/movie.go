package model

type Movie struct {
	ID          string   `json:"id" bson:"id"`
	Name        string   `json:"name" bson:"name"`
	Year        string   `json:"year" bson:"year"`
	Rating      string   `json:"rating" bson:"rating"`
	RatingCount string   `json:"rating_count" bson:"rating_count"`
	Duration    string   `json:"duration" bson:"duration"`
	Genres      []string `json:"genres" bson:"genres"`
	Poster      string   `json:"poster" bson:"poster"`
	Budget      string   `json:"budget" bson:"budget"`
	Cumulative  string   `json:"cumulative" bson:"cumulative"`
	Director    string   `json:"director" bson:"director"`
	Stars       []string `json:"stars" bson:"stars"`
	Country     []string `json:"country" bson:"country"`
	StoryLine   string   `json:"story_line" bson:"story_line"`
}
