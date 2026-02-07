package main

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	userCacheKeyPrefix = "user:info:"
	userCacheTTL       = 24 * time.Hour
)

// UserCacheRepository caches user data in Redis (TTL 1 day).
type UserCacheRepository struct {
	client *redis.Client
	ctx    context.Context
}

// NewUserCacheRepository creates a new user cache repository.
func NewUserCacheRepository(client *redis.Client) *UserCacheRepository {
	return &UserCacheRepository{
		client: client,
		ctx:    context.Background(),
	}
}

// Get returns the cached user, or (nil, nil) if not found.
func (r *UserCacheRepository) Get(userID int) (*User, error) {
	key := userCacheKeyPrefix + strconv.Itoa(userID)
	data, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Set stores the user in cache with 24h TTL.
func (r *UserCacheRepository) Set(userID int, user *User) error {
	key := userCacheKeyPrefix + strconv.Itoa(userID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, userCacheTTL).Err()
}
