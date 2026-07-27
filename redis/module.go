package redis

import (
	core "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app"
)

// Module wires a general-purpose Redis client into the fx application: it supplies
// the given Config and provides NewService. The constructor comes first, the
// config is explicit; modes are variadic (empty = always registered).
//
//	redis.Module(&cfg.Redis, engine.Scheduler, engine.Batch)
func Module(config *Config, modes ...string) {
	core.Supply(config, modes...)
	core.Provide(NewService, modes...)
}
