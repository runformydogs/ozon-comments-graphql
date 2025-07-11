package storage_test

import (
	"context"
	"ozon-comments-graphql/internal/models"
	"ozon-comments-graphql/internal/storage"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_Posts(t *testing.T) {
	s := storage.NewMemoryStorage()
	ctx := context.Background()

	post := s.CreatePost(ctx, "Test Post", "Test Content")
	assert.NotEmpty(t, post.ID)
	assert.Equal(t, "Test Post", post.Title)

	posts := s.ListPosts(ctx)
	assert.Len(t, posts, 1)
	assert.Equal(t, post.ID, posts[0].ID)

	foundPost, err := s.GetPost(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, post.ID, foundPost.ID)
}

func TestMemoryStorage_Comments(t *testing.T) {
	s := storage.NewMemoryStorage()
	ctx := context.Background()

	post := s.CreatePost(ctx, "Test Post", "Test Content")

	comment, err := s.CreateComment(ctx, post.ID, nil, "Test Comment")
	assert.NoError(t, err)
	assert.NotEmpty(t, comment.ID)
	assert.Equal(t, "Test Comment", comment.Content)

	comments, next := s.ListComments(ctx, post.ID, 10, nil)
	assert.Len(t, comments, 1)
	assert.Equal(t, comment.ID, comments[0].ID)
	assert.Nil(t, next)

	childComment, err := s.CreateComment(ctx, post.ID, &comment.ID, "Child Comment")
	assert.NoError(t, err)

	comments, _ = s.ListComments(ctx, post.ID, 10, nil)
	assert.Len(t, comments, 2)
	assert.Equal(t, comment.ID, *childComment.ParentID)
}

func TestMemoryStorage_CommentErrors(t *testing.T) {
	s := storage.NewMemoryStorage()
	ctx := context.Background()

	post := s.CreatePost(ctx, "Test Post", "Test Content")

	longText := string(make([]byte, 2001))
	_, err := s.CreateComment(ctx, post.ID, nil, longText)
	assert.ErrorIs(t, err, storage.ErrTooLong)

	_, err = s.ToggleComments(ctx, post.ID, true)
	assert.NoError(t, err)

	_, err = s.CreateComment(ctx, post.ID, nil, "Should fail")
	assert.ErrorIs(t, err, storage.ErrForbidden)
}

func TestMemoryStorage_Pagination(t *testing.T) {
	s := storage.NewMemoryStorage()
	ctx := context.Background()

	post := s.CreatePost(ctx, "Test Post", "Test Content")

	var comments []*models.Comment
	for i := 0; i < 15; i++ {
		c, err := s.CreateComment(ctx, post.ID, nil, "Comment")
		assert.NoError(t, err)
		comments = append(comments, c)
	}

	page1, next1 := s.ListComments(ctx, post.ID, 5, nil)
	assert.Len(t, page1, 5)
	assert.NotNil(t, next1)
	assert.Equal(t, comments[4].ID, *next1)

	page2, next2 := s.ListComments(ctx, post.ID, 5, next1)
	assert.Len(t, page2, 5)
	assert.NotNil(t, next2)
	assert.Equal(t, comments[9].ID, *next2)

	page3, next3 := s.ListComments(ctx, post.ID, 5, next2)
	assert.Len(t, page3, 5)
	assert.Nil(t, next3)
}
