package graph

import (
	"foodworks.ml/m/ent"
	"github.com/go-redis/redis/v8"
)

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Client *ent.Client
	Redis  *redis.Client
}
