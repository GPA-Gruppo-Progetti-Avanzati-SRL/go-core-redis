# Codici di errore — go-core-redis

**Il modulo non definisce nessun codice di errore applicativo**: non produce
`*core.ApplicationError` e non ha una tabella di codici. Gli errori di `redis.Service`
risalgono così come li ritorna il client `go-redis` (incluso `redis.Nil` per una chiave
assente); chi li espone via HTTP li avvolge nel proprio `core.TechnicalError().WithCause(err)`,
ottenendo `TECH500`.

## Errori sentinella del locker

`locker/` implementa `lock.Locker` su redsync (Redlock) e ritorna i due errori neutri di
`go-core-app/lock`:

| Errore | Origine | Significato |
|---|---|---|
| `lock.ErrNotAcquired` | `locker/locker.go:59` | il lock è già tenuto da un'altra replica. È **contesa, non un guasto**: il chiamante (lo scheduler, via `scheduler/gocronlock`) salta il tick. Il claiming sul DB garantisce comunque la correttezza |
| `lock.ErrLockLost` | `locker/locker.go:92,97` | il rinnovo del lease (`Extend`) è fallito o ha ritornato false: il TTL redsync (~30s di default) è scaduto e un'altra replica può aver preso il lock. Il lavoro in corso non è più protetto |

Entrambi sono avvolti con `fmt.Errorf("redis lock extend %q: %w", ...)`: vanno confrontati con
`errors.Is`, mai con `==`.

## Errori di configurazione

Config non valida (`validate:` sulla `redis.Config`) o connessione impossibile all'avvio →
l'errore risale a fx e **l'app non parte**. Non c'è codice: non c'è chiamante da informare.
