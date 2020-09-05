package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/jwtauth"

	"foodworks.ml/m/auth"
	"foodworks.ml/m/internal/generated/ent"
	generated "foodworks.ml/m/internal/generated/graphql"
	"foodworks.ml/m/internal/resolver"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/facebook/ent/dialect"
	entsql "github.com/facebook/ent/dialect/sql"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/cors"
)

func Server(es graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(es)
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		KeepAlivePingInterval: 15 * time.Second,
	})

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	return srv
}

// Open new db connection
func Open(databaseURL string) (*sql.DB, *ent.Client) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)
	return db, ent.NewClient(ent.Driver(drv))
}

func main() {
	// init db client
	db, client := Open(os.Getenv("POSTGRES_URL"))

	defer client.Close()
	defer db.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Printf("failed creating schema resources: %v", err)
		return
	}
	// init redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"), // no password set
		DB:       0,                       // use default DB
	})

	// init server
	router := chi.NewRouter()
	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		// AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		// Debug:            true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}).Handler)
	router.Use(jwtauth.Verifier(auth.TokenAuth))
	router.Use(auth.Middleware())

	srv := Server(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{Client: client, Redis: rdb}}))
	appPort := os.Getenv("APPLICATION_PORT")

	router.Handle("/graphql/playground", playground.Handler("GraphQL playground", "/graphql"))
	router.Handle("/graphql", srv)
	router.Post("/auth", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Hi")
		// A very simple health check.
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
	})
	router.Post("/.api/auth", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Hi")
		// A very simple health check.
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
	})
	router.Post("/*", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Hi")
		// A very simple health check.
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
	})

	log.Printf("connect to http://localhost:%s/graphql/playground for GraphQL playground", appPort)
	err := http.ListenAndServe(appPort, router)
	if err != nil {
		panic(err)
	}
}
