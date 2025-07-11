package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"ozon-comments-graphql/graph"
	"ozon-comments-graphql/internal/storage"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not loaded: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	var store storage.Storage
	if os.Getenv("STORAGE_TYPE") == "postgres" {
		pgStore, err := storage.NewPostgresStorage(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatal("Postgres init failed:", err)
		}
		store = pgStore
	} else {
		store = storage.NewMemoryStorage()
	}

	resolver := &graph.Resolver{
		Store: store,
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
