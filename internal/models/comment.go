package models

import "time"

type Comment struct {
	ID        string
	PostID    string
	ParentID  *string
	Content   string
	CreatedAt time.Time
}
