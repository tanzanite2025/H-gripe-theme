package app

import (
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/service"

	"gorm.io/gorm"
)

type dependencyServicesBuilder struct {
	db                           *gorm.DB
	redisCache                   *cache.RedisCache
	cfg                          *config.Config
	repos                        Repositories
	support                      *dependencySupport
	merchantOutboxPublisher      *service.MerchantOutboxPublisher
	productCacheOutboxPublisher  *service.ProductCacheOutboxPublisher
	services                     Services
	customerServiceRealtimeRelay *service.CustomerServiceRealtimeRelay
}

func newDependencyServices(
	db *gorm.DB,
	redisCache *cache.RedisCache,
	cfg *config.Config,
	repos Repositories,
	support *dependencySupport,
) (Services, *service.CustomerServiceRealtimeRelay, error) {
	builder := &dependencyServicesBuilder{
		db:         db,
		redisCache: redisCache,
		cfg:        cfg,
		repos:      repos,
		support:    support,
	}

	if err := builder.build(); err != nil {
		return Services{}, nil, err
	}
	if err := builder.wire(); err != nil {
		return Services{}, nil, err
	}

	return builder.services, builder.customerServiceRealtimeRelay, nil
}
