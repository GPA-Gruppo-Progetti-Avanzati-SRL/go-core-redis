package locker

import (
	core "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app/lock"
)

// Module registers the Redis-backed lock.Locker in the fx application. It is
// modes-only (config-free): it consumes the *redis.Client provided by
// redis.Module, so wire redis.Module first.
//
//	redis.Module(&cfg.Redis, engine.Scheduler, engine.Batch)
//	batch.Module(&cfg.Batch, batch.WithLocker(locker.Module), ...)
func Module(modes ...string) {
	core.ProvideAs[lock.Locker](New, modes...)
}
