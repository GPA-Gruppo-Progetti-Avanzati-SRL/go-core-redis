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

## Lock distribuito — `locker`

Implementa `lock.Locker` di go-core-app su redsync (Redlock). Non conosce gocron né lo scheduler:
è `go-core-batch` ad adattare un `lock.Locker` a gocron, internamente.

```go
import (
    "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-redis/redis"
    redislocker "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-redis/locker"
)

redis.Module(&svc.Redis, engine.Scheduler, engine.Batch)   // prima il client

batch.Module(&svc.Batch, Register,
    batch.WithStore(storemongo.Module),
    batch.WithLocker(redislocker.Module),                  // poi il locker
    // ...
)
```

`locker.Module(modes ...string)` è **modes-only**: non ha config, consuma il `*goredis.Client`
fornito da `redis.Module`, che va quindi wirato prima. Registra `lock.Locker` con
`core.ProvideAs`.

### Uso diretto

```go
h, err := locker.Acquire(ctx, "import-anagrafiche")
if errors.Is(err, lock.ErrNotAcquired) {
    return nil          // qualcun altro sta già lavorando: si salta il giro
}
if err != nil {
    return err
}
defer h.Release(ctx)
```

Senza opzioni `Acquire` fa **un solo tentativo non bloccante** con TTL di **30s** e ritorna
`lock.ErrNotAcquired` in contesa: è la semantica dispatch-dedup su cui si appoggia lo scheduler
batch. Per una sezione critica che richiede mutua esclusione:

```go
h, err := locker.Acquire(ctx, "chiave",
    lock.WithWait(2*time.Minute, 200*time.Millisecond),   // blocca e ritenta
    lock.WithExpiry(5*time.Minute))
```

`Handle.Extend(ctx)` rinnova il TTL di una sezione critica lunga, e ritorna `lock.ErrLockLost` se
il lock è nel frattempo scaduto ed è stato rubato — a differenza di `Release`, dove un lock già
scaduto è benigno per un lock di dedup.

> **Il lock dello scheduler batch è un'ottimizzazione, non correttezza.** La correttezza del batch
> è garantita dal DB claiming (`store.ClaimBatch`); il lock evita solo che N repliche eseguano lo
> stesso tick cron. Un'app mongo-only o sql-only può usare
> [`go-core-mongo/locker`](../go-core-mongo) o [`go-core-sql/locker`](../go-core-sql) e **non
> deployare Redis**.

---

## Comandi

```bash
go build ./...
go test ./...
go test -race -count=2 ./...
go vet ./...
```
