package storage

import (
	"context"
	"errors"
	"ozon-comments-graphql/internal/models"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("comments disabled")
	ErrTooLong    = errors.New("comment too long")
	maxCommentLen = 2000
)

type MemoryStorage struct {
	mu       sync.RWMutex
	posts    map[string]*models.Post
	comments map[string]*models.Comment
	byPost   map[string][]*models.Comment
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		posts:    make(map[string]*models.Post),
		comments: make(map[string]*models.Comment),
		byPost:   make(map[string][]*models.Comment),
	}
}

func (s *MemoryStorage) CreatePost(_ context.Context, title, content string) *models.Post {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := &models.Post{
		ID:        uuid.NewString(),
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
	}
	s.posts[p.ID] = p
	return p
}

func (s *MemoryStorage) ToggleComments(_ context.Context, id string, d bool) (*models.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.posts[id]
	if !ok {
		return nil, ErrNotFound
	}
	p.CommentsDisabled = d
	return p, nil
}

func (s *MemoryStorage) ListPosts(_ context.Context) []*models.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*models.Post, 0, len(s.posts))
	for _, p := range s.posts {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStorage) GetPost(_ context.Context, id string) (*models.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.posts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *MemoryStorage) CreateComment(_ context.Context, postID string, parentID *string, text string) (*models.Comment, error) {
	if len(text) > maxCommentLen {
		return nil, ErrTooLong
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.posts[postID]
	if !ok {
		return nil, ErrNotFound
	}
	if p.CommentsDisabled {
		return nil, ErrForbidden
	}

	c := &models.Comment{
		ID:        uuid.NewString(),
		PostID:    postID,
		ParentID:  parentID,
		Content:   text,
		CreatedAt: time.Now(),
	}
	s.comments[c.ID] = c
	s.byPost[postID] = append(s.byPost[postID], c)
	return c, nil
}

func (s *MemoryStorage) ListComments(_ context.Context, postID string, first int, afterID *string) ([]*models.Comment, *string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := s.byPost[postID]
	if len(all) == 0 {
		return nil, nil
	}

	start := 0
	if afterID != nil {
		for i, c := range all {
			if c.ID == *afterID {
				start = i + 1
				break
			}
		}
	}

	end := start + first
	if end > len(all) {
		end = len(all)
	}

	items := all[start:end]
	var next *string
	if end < len(all) {
		next = &items[len(items)-1].ID
	}
	return items, next
}
