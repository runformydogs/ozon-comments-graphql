package storage

import (
	"context"
	"ozon-comments-graphql/internal/models"
)

type Storage interface {
	CreatePost(ctx context.Context, title, content string) *models.Post
	ToggleComments(ctx context.Context, id string, disabled bool) (*models.Post, error)
	ListPosts(ctx context.Context) []*models.Post
	GetPost(ctx context.Context, id string) (*models.Post, error)
	CreateComment(ctx context.Context, postID string, parentID *string, content string) (*models.Comment, error)
	ListComments(ctx context.Context, postID string, first int, afterID *string) ([]*models.Comment, *string)
}
