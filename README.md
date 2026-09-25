# GO-CORE-REDIS

## Installation

    go get github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-redis

---

Client Redis general-purpose per le applicazioni GPA, più un `lock.Locker` Redis-backed. Due
package indipendenti — il modulo **non ha package a root**:

| Package | Contenuto |
|---|---|
| `redis/` | client [go-redis/v9](https://redis.uptrace.dev/) wirato su fx (`*goredis.Client`) |
| `locker/` | implementazione Redis di [`lock.Locker`](../go-core-app) su [redsync](https://github.com/go-redsync/redsync) (Redlock) |

Dipende da [`go-core-app`](../go-core-app). Nessuna dipendenza da gocron o da go-core-batch.
**Richiede Go 1.27+.**

---

## Client — `redis`

`redis.Module(cfg *Config, modes ...string)` supplisce la config e fornisce il `*goredis.Client` a
fx. La disconnessione è agganciata al lifecycle (`OnStop`).

```go
import "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-redis/redis"

func main() {
    svc := core.Boot[app.Config, services.Config](core.App{ /* ... */ })

    redis.Module(&svc.Redis, engine.Api, engine.Worker)

    core.Run()
}
```

Il client è un `*goredis.Client` non incapsulato: si inietta e si usa come tale.

```go
type data struct {
    core.In
    Redis *goredis.Client
}
```

```yaml
config:
  services:
    redis:
      enable: true
      address: redis-master
      port: 6379
      password: ${REDIS_PWD}
```

`redis.Module` usa `core.Module("redis", ...)`: il `*Config` è **privato** al modulo, il
`*goredis.Client` è esportato al grafo dell'app.

---

## Lock distribuito

Non è più qui: il backend Redis del lock è **`go-core-locker/redisstore`**, che consuma il
`*goredis.Client` fornito da `redis.Module`. Con lui è uscita anche la dipendenza `redsync`: con un
solo client non eseguiva Redlock ma un `SET NX`.

```go
redis.Module(&cfg.Redis, engine.Scheduler)
corelock.Module(&cfg.Lock, corelock.WithBackend(redisstore.Module))
```

## Comandi

```bash
go build ./...
go test ./...
go test -race -count=2 ./...
go vet ./...
```
