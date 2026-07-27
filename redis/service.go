package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
)

// NewService builds a general-purpose go-redis client and registers an fx OnStop
// hook that closes it. The returned *goredis.Client can be used for any Redis
// need (caching, pub/sub, distributed locking via the locker subpackage, ...).
func NewService(config *Config, lc fx.Lifecycle) *goredis.Client {

	redisOptions := &goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Address, config.Port),
		Password: config.Password,
	}
	redisClient := goredis.NewClient(redisOptions)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if redisClient != nil {
				log.Info().Msg("Disconnecting redis")
				if err := redisClient.Close(); err != nil {
					log.Error().Err(err).Msg("Failed to disconnect Redis")
					return err
				}
			}
			return nil
		}})

	return redisClient
}
