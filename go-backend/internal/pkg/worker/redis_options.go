package worker

import (
	"commerce-platform/internal/pkg/config"

	"github.com/hibiken/asynq"
)

// newRedisConnOpt keeps Asynq's client and server on the same Redis
// high-availability path as the rest of the application.
func newRedisConnOpt(cfg *config.RedisConfig) asynq.RedisConnOpt {
	if cfg.NormalizedMode() == "sentinel" {
		return asynq.RedisFailoverClientOpt{
			MasterName:       cfg.MasterName,
			SentinelAddrs:    cfg.GetRedisAddrs(),
			SentinelUsername: cfg.SentinelUsername,
			SentinelPassword: cfg.SentinelPassword,
			Username:         cfg.Username,
			Password:         cfg.Password,
			DB:               cfg.DB,
			PoolSize:         cfg.PoolSize,
		}
	}

	addr := cfg.GetRedisAddr()
	if addrs := cfg.GetRedisAddrs(); len(addrs) > 0 {
		addr = addrs[0]
	}

	return asynq.RedisClientOpt{
		Addr:     addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	}
}
