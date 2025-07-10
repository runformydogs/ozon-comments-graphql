package graph

import (
	"ozon-comments-graphql/internal/storage"
)

type Resolver struct {
	Store storage.Storage
}
