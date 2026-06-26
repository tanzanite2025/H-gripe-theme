package worker

import (
	"context"
	"tanzanite/internal/pkg/config"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"tanzanite/internal/pkg/logger"
)

// Server encapsulates the asynq worker server
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewServer initializes a new Asynq worker server
func NewServer(cfg *config.RedisConfig) *Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("process task failed", zap.String("type", task.Type()), zap.Error(err))
			}),
		},
	)

	mux := asynq.NewServeMux()
	// Register handlers here if needed:
	// mux.HandleFunc("task:type", handlerFunc)

	return &Server{
		server: srv,
		mux:    mux,
	}
}

// Start runs the worker server
func (s *Server) Start() error {
	logger.Info("Starting Asynq worker server")
	return s.server.Start(s.mux)
}

// Stop gracefully shuts down the worker server
func (s *Server) Stop() {
	logger.Info("Stopping Asynq worker server")
	s.server.Shutdown()
}

// Mux returns the underlying ServeMux so handlers can be registered externally
func (s *Server) Mux() *asynq.ServeMux {
	return s.mux
}
