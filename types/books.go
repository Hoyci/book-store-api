package types

import "time"

type Book struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Author        string    `json:"author"`
	Genres        []string  `json:"genres"`
	ReleaseYear   int32     `json:"releaseYear"`
	NumberOfPages int32     `json:"numberOfPages"`
	ImageUrl      string    `json:"imageUrl"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CreateBookPayload struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Author        string   `json:"author"`
	Genres        []string `json:"genres"`
	ReleaseYear   int32    `json:"releaseYear"`
	NumberOfPages int32    `json:"numberOfPages"`
	ImageUrl      string   `json:"imageUrl"`
}
