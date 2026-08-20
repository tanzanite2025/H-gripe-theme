package worker

import (
	"testing"

	"commerce-platform/internal/pkg/config"

	"github.com/hibiken/asynq"
)

func TestNewRedisConnOptStandalone(t *testing.T) {
	cfg := &config.RedisConfig{
		Mode:     "standalone",
		Addrs:    []string{"redis.internal:6380"},
		Host:     "localhost",
		Port:     9510,
		Username: "app",
		Password: "secret",
		DB:       4,
		PoolSize: 24,
	}

	opt, ok := newRedisConnOpt(cfg).(asynq.RedisClientOpt)
	if !ok {
		t.Fatalf("newRedisConnOpt() type = %T, want asynq.RedisClientOpt", newRedisConnOpt(cfg))
	}
	if opt.Addr != "redis.internal:6380" {
		t.Fatalf("Addr = %q, want %q", opt.Addr, "redis.internal:6380")
	}
	if opt.Username != cfg.Username || opt.Password != cfg.Password || opt.DB != cfg.DB || opt.PoolSize != cfg.PoolSize {
		t.Fatalf("standalone Redis options did not preserve credentials/database/pool size: %+v", opt)
	}
}

func TestNewRedisConnOptSentinel(t *testing.T) {
	cfg := &config.RedisConfig{
		Mode:             "sentinel",
		Addrs:            []string{"sentinel-1:26379", "sentinel-2:26379"},
		MasterName:       "mymaster",
		SentinelUsername: "sentinel-user",
		SentinelPassword: "sentinel-secret",
		Username:         "redis-user",
		Password:         "redis-secret",
		DB:               2,
		PoolSize:         18,
	}

	opt, ok := newRedisConnOpt(cfg).(asynq.RedisFailoverClientOpt)
	if !ok {
		t.Fatalf("newRedisConnOpt() type = %T, want asynq.RedisFailoverClientOpt", newRedisConnOpt(cfg))
	}
	if opt.MasterName != cfg.MasterName {
		t.Fatalf("MasterName = %q, want %q", opt.MasterName, cfg.MasterName)
	}
	if len(opt.SentinelAddrs) != len(cfg.Addrs) || opt.SentinelAddrs[0] != cfg.Addrs[0] || opt.SentinelAddrs[1] != cfg.Addrs[1] {
		t.Fatalf("SentinelAddrs = %#v, want %#v", opt.SentinelAddrs, cfg.Addrs)
	}
	if opt.SentinelUsername != cfg.SentinelUsername || opt.SentinelPassword != cfg.SentinelPassword {
		t.Fatalf("sentinel credentials did not survive conversion: %+v", opt)
	}
	if opt.Username != cfg.Username || opt.Password != cfg.Password || opt.DB != cfg.DB || opt.PoolSize != cfg.PoolSize {
		t.Fatalf("Redis options did not preserve credentials/database/pool size: %+v", opt)
	}
}
