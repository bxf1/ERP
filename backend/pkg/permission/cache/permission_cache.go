package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type PermissionCache struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

func NewPermissionCache(redisURL string, ttl time.Duration) (*PermissionCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &PermissionCache{
		client: client,
		ttl:    ttl,
		prefix: "rbac:",
	}, nil
}

type CachedUserPermissions struct {
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}

func (c *PermissionCache) key(parts ...string) string {
	k := c.prefix
	for i, p := range parts {
		if i > 0 {
			k += ":"
		}
		k += p
	}
	return k
}

// GetUserPermissions returns cached permissions for a user.
func (c *PermissionCache) GetUserPermissions(ctx context.Context, userID string) (*CachedUserPermissions, error) {
	data, err := c.client.Get(ctx, c.key("user", userID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var perms CachedUserPermissions
	if err := json.Unmarshal(data, &perms); err != nil {
		return nil, err
	}
	return &perms, nil
}

// SetUserPermissions caches permissions for a user.
func (c *PermissionCache) SetUserPermissions(ctx context.Context, userID string, perms *CachedUserPermissions) error {
	data, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key("user", userID), data, c.ttl).Err()
}

// InvalidateUser removes cached permissions for a specific user.
func (c *PermissionCache) InvalidateUser(ctx context.Context, userID string) error {
	return c.client.Del(ctx, c.key("user", userID)).Err()
}

// InvalidateRole removes cached permissions for all users with the given role.
func (c *PermissionCache) InvalidateRole(ctx context.Context, roleCode string) error {
	iter := c.client.Scan(ctx, 0, c.key("user", "*"), 0).Iterator()
	for iter.Next(ctx) {
		data, err := c.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var perms CachedUserPermissions
		if err := json.Unmarshal(data, &perms); err != nil {
			continue
		}
		for _, r := range perms.Roles {
			if r == roleCode {
				if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
					return err
				}
				break
			}
		}
	}
	return iter.Err()
}

// InvalidateAll clears the entire permission cache.
func (c *PermissionCache) InvalidateAll(ctx context.Context) error {
	iter := c.client.Scan(ctx, 0, c.key("*"), 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

// InvalidatePermissions removes all user cache entries (used when permission definitions change).
func (c *PermissionCache) InvalidatePermissions(ctx context.Context) error {
	return c.InvalidateAll(ctx)
}
