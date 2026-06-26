package worker

import (
	"tanzanite/internal/pkg/config"

	"github.com/hibiken/asynq"
)

// Client encapsulates the asynq client
type Client struct {
	client *asynq.Client
}

// NewClient initializes a new Asynq client for enqueuing tasks
func NewClient(cfg *config.RedisConfig) *Client {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := asynq.NewClient(redisOpt)

	return &Client{
		client: client,
	}
}

// Enqueue wraps the underlying asynq Enqueue method
func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, opts...)
}

// Close closes the underlying asynq client
func (c *Client) Close() error {
	return c.client.Close()
}
