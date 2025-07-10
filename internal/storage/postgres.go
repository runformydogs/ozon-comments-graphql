package storage

import (
	"context"
	"errors"
	"fmt"
	"ozon-comments-graphql/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, connString string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	if err := createTables(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PostgresStorage{db: pool}, nil
}

func createTables(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS posts (
			id UUID PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			comments_disabled BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS comments (
			id UUID PRIMARY KEY,
			post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
			parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		);
		
		CREATE INDEX IF NOT EXISTS comments_post_id_idx ON comments (post_id);
		CREATE INDEX IF NOT EXISTS comments_created_at_idx ON comments (created_at);
	`)
	return err
}

func (s *PostgresStorage) CreatePost(ctx context.Context, title, content string) *models.Post {
	id := uuid.NewString()
	now := time.Now()

	_, err := s.db.Exec(ctx,
		`INSERT INTO posts (id, title, content, comments_disabled, created_at) VALUES ($1, $2, $3, $4, $5)`,
		id, title, content, false, now,
	)
	if err != nil {
		fmt.Printf("CreatePost error: %v\n", err)
	}

	return &models.Post{
		ID:               id,
		Title:            title,
		Content:          content,
		CommentsDisabled: false,
		CreatedAt:        now,
	}
}

func (s *PostgresStorage) ToggleComments(ctx context.Context, id string, disabled bool) (*models.Post, error) {
	_, err := s.db.Exec(ctx, "UPDATE posts SET comments_disabled = $1 WHERE id = $2", disabled, id)
	if err != nil {
		return nil, err
	}

	return s.GetPost(ctx, id)
}

func (s *PostgresStorage) ListPosts(ctx context.Context) []*models.Post {
	rows, err := s.db.Query(ctx, "SELECT id, title, content, comments_disabled, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CommentsDisabled, &p.CreatedAt); err != nil {
			continue
		}
		posts = append(posts, &p)
	}

	return posts
}

func (s *PostgresStorage) GetPost(ctx context.Context, id string) (*models.Post, error) {
	var p models.Post
	err := s.db.QueryRow(ctx, "SELECT id, title, content, comments_disabled, created_at FROM posts WHERE id = $1", id).Scan(
		&p.ID, &p.Title, &p.Content, &p.CommentsDisabled, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *PostgresStorage) CreateComment(ctx context.Context, postID string, parentID *string, content string) (*models.Comment, error) {
	if len(content) > maxCommentLen {
		return nil, ErrTooLong
	}

	var commentsDisabled bool
	err := s.db.QueryRow(ctx, "SELECT comments_disabled FROM posts WHERE id = $1", postID).Scan(&commentsDisabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if commentsDisabled {
		return nil, ErrForbidden
	}

	id := uuid.NewString()
	now := time.Now()

	_, err = s.db.Exec(ctx,
		`INSERT INTO comments (id, post_id, parent_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`,
		id, postID, parentID, content, now,
	)
	if err != nil {
		return nil, err
	}

	return &models.Comment{
		ID:        id,
		PostID:    postID,
		ParentID:  parentID,
		Content:   content,
		CreatedAt: now,
	}, nil
}

func (s *PostgresStorage) ListComments(ctx context.Context, postID string, first int, afterID *string) ([]*models.Comment, *string) {
	query := `SELECT id, post_id, parent_id, content, created_at 
              FROM comments 
              WHERE post_id = $1 `

	params := []interface{}{postID}
	order := "ORDER BY created_at ASC"

	if afterID != nil {
		query += " AND created_at > (SELECT created_at FROM comments WHERE id = $2) "
		params = append(params, *afterID)
		order = "ORDER BY created_at ASC"
	}

	query += order + " LIMIT $2"
	if afterID != nil {
		params = append(params, first)
	} else {
		params = append(params, first)
	}

	rows, err := s.db.Query(ctx, query, params...)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var c models.Comment
		var parentID *string
		if err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.Content, &c.CreatedAt); err != nil {
			continue
		}
		c.ParentID = parentID
		comments = append(comments, &c)
	}

	var nextCursor *string
	if len(comments) > 0 {
		lastID := comments[len(comments)-1].ID
		nextCursor = &lastID
	}

	return comments, nextCursor
}
