package redis

import (
	core "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app"
)

// Module wires a general-purpose Redis client into the fx application: it supplies
// the given Config and provides NewService. The constructor comes first, the
// config is explicit; modes are variadic (empty = always registered).
//
// Redis è un driver: le registrazioni stanno in un core.Module("redis"), quindi la *Config è
// privata al modulo (non iniettabile dal grafo dell'app) mentre il *Service resta esportato —
// è l'handle che l'app (e locker.Module) consuma.
//
//	redis.Module(&cfg.Redis, engine.Scheduler, engine.Batch)
func Module(config *Config, modes ...string) {
	core.Module("redis", func() {
		core.Supply(config, modes...)
		core.Provide(NewService, modes...)
	})
}
