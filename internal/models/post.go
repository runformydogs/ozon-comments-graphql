package models

import "time"

type Post struct {
	ID               string
	Title            string
	Content          string
	CommentsDisabled bool
	CreatedAt        time.Time
}
