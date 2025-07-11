package graph_test

import (
	"context"
	"testing"
	"time"

	"ozon-comments-graphql/graph"
	"ozon-comments-graphql/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestPostCreation(t *testing.T) {
	r := &graph.Resolver{
		Store:  storage.NewMemoryStorage(),
		Broker: graph.NewCommentBroker(),
	}

	post, err := r.Mutation().CreatePost(context.Background(), "Title", "Content")
	assert.NoError(t, err)
	assert.Equal(t, "Title", post.Title)
	assert.False(t, post.CommentsDisabled)
}

func TestCommentWorkflow(t *testing.T) {
	r := &graph.Resolver{
		Store:  storage.NewMemoryStorage(),
		Broker: graph.NewCommentBroker(),
	}
	ctx := context.Background()

	post, _ := r.Mutation().CreatePost(ctx, "Test", "Content")

	comment, err := r.Mutation().CreateComment(ctx, post.ID, nil, "My comment")
	assert.NoError(t, err)
	assert.Equal(t, "My comment", comment.Content)

	comments, err := r.Query().Comments(ctx, post.ID, nil, nil)
	assert.NoError(t, err)
	assert.Len(t, comments.Items, 1)
}

func TestSubscriptionSimple(t *testing.T) {
	r := &graph.Resolver{
		Store:  storage.NewMemoryStorage(),
		Broker: graph.NewCommentBroker(),
	}
	ctx := context.Background()

	post, _ := r.Mutation().CreatePost(ctx, "Sub test", "Content")

	subCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	ch, err := r.Subscription().CommentAdded(subCtx, post.ID)
	assert.NoError(t, err)

	newComment, _ := r.Mutation().CreateComment(subCtx, post.ID, nil, "New!")

	select {
	case msg := <-ch:
		assert.Equal(t, newComment.ID, msg.ID)
	case <-time.After(500 * time.Millisecond):
		assert.Fail(t, "Не получили комментарий по подписке")
	}
}

func TestDisabledComments(t *testing.T) {
	r := &graph.Resolver{
		Store: storage.NewMemoryStorage(),
	}
	ctx := context.Background()

	post, _ := r.Mutation().CreatePost(ctx, "Test", "Content")

	_, err := r.Mutation().ToggleComments(ctx, post.ID, true)
	assert.NoError(t, err)

	_, err = r.Mutation().CreateComment(ctx, post.ID, nil, "Test")
	assert.Error(t, err)
}
