package graph

import (
	"ozon-comments-graphql/graph/model"
	"sync"
)

type CommentBroker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan *model.Comment]struct{}
}

func NewCommentBroker() *CommentBroker {
	return &CommentBroker{
		subscribers: make(map[string]map[chan *model.Comment]struct{}),
	}
}

func (b *CommentBroker) Subscribe(postID string) chan *model.Comment {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan *model.Comment, 1)

	if _, ok := b.subscribers[postID]; !ok {
		b.subscribers[postID] = make(map[chan *model.Comment]struct{})
	}
	b.subscribers[postID][ch] = struct{}{}

	return ch
}

func (b *CommentBroker) Unsubscribe(postID string, ch chan *model.Comment) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subscribers[postID]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(b.subscribers, postID)
		}
	}
}

func (b *CommentBroker) Publish(comment *model.Comment) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	postID := comment.PostID
	if subs, ok := b.subscribers[postID]; ok {
		for ch := range subs {
			select {
			case ch <- comment:
			default:
			}
		}
	}
}
