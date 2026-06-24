package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(redisUrl string) (*Client, error) {
	opt, _ := redis.ParseURL(redisUrl)
	rdb := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Client{
		rdb: rdb,
	}, nil
}

func (c *Client) GetLatestID(ctx context.Context) (string, error) {
	id, err := c.rdb.Get(ctx, "scraper:latest_post_id").Result()
	if err == redis.Nil {
		return "", nil
	}
	return id, err
}

func (c *Client) SetLatestID(ctx context.Context, latestPostID string) error {
	return c.rdb.Set(ctx, "scraper:latest_post_id", latestPostID, 0).Err()
}
