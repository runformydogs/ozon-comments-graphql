package graph_test

import (
	"context"
	"testing"

	"ozon-comments-graphql/graph"
	"ozon-comments-graphql/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestResolvers_Post(t *testing.T) {
	resolver := &graph.Resolver{
		Store: storage.NewMemoryStorage(),
	}
	ctx := context.Background()

	post, err := resolver.Mutation().CreatePost(ctx, "Test", "Content")
	assert.NoError(t, err)
	assert.Equal(t, "Test", post.Title)

	foundPost, err := resolver.Query().Post(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, post.ID, foundPost.ID)
}

func TestResolvers_Comments(t *testing.T) {
	resolver := &graph.Resolver{
		Store: storage.NewMemoryStorage(),
	}
	ctx := context.Background()

	post, _ := resolver.Mutation().CreatePost(ctx, "Test", "Content")

	comment, err := resolver.Mutation().CreateComment(ctx, post.ID, nil, "Comment text")
	assert.NoError(t, err)
	assert.Equal(t, "Comment text", comment.Content)

	commentPage, err := resolver.Query().Comments(ctx, post.ID, nil, nil)
	assert.NoError(t, err)
	assert.Len(t, commentPage.Items, 1)
	assert.Equal(t, comment.ID, commentPage.Items[0].ID)
}

func TestResolvers_ToggleComments(t *testing.T) {
	resolver := &graph.Resolver{
		Store: storage.NewMemoryStorage(),
	}
	ctx := context.Background()

	post, _ := resolver.Mutation().CreatePost(ctx, "Test", "Content")

	updatedPost, err := resolver.Mutation().ToggleComments(ctx, post.ID, true)
	assert.NoError(t, err)
	assert.True(t, updatedPost.CommentsDisabled)

	_, err = resolver.Mutation().CreateComment(ctx, post.ID, nil, "Should fail")
	assert.Error(t, err)
}

func TestResolvers_Pagination(t *testing.T) {
	resolver := &graph.Resolver{
		Store: storage.NewMemoryStorage(),
	}
	ctx := context.Background()

	post, _ := resolver.Mutation().CreatePost(ctx, "Test", "Content")

	for i := 0; i < 15; i++ {
		_, err := resolver.Mutation().CreateComment(ctx, post.ID, nil, "Comment")
		assert.NoError(t, err)
	}

	first := int32(5)
	page1, err := resolver.Query().Comments(ctx, post.ID, &first, nil)
	assert.NoError(t, err)
	assert.Len(t, page1.Items, 5)
	assert.NotNil(t, page1.NextCursor)

	page2, err := resolver.Query().Comments(ctx, post.ID, &first, page1.NextCursor)
	assert.NoError(t, err)
	assert.Len(t, page2.Items, 5)
}
