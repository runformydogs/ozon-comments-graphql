package graph_test

import (
	"context"
	"ozon-comments-graphql/graph"
	"ozon-comments-graphql/internal/storage"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptions(t *testing.T) {
	store := storage.NewMemoryStorage()
	broker := graph.NewCommentBroker()
	resolver := &graph.Resolver{
		Store:  store,
		Broker: broker,
	}

	post := store.CreatePost(context.Background(), "Test", "Content")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	subCh, err := resolver.Subscription().CommentAdded(ctx, post.ID)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		comment := <-subCh
		assert.Equal(t, "Test comment", comment.Content)
	}()

	_, err = resolver.Mutation().CreateComment(ctx, post.ID, nil, "Test comment")
	assert.NoError(t, err)

	wg.Wait()
}
