# Codici di errore — go-core-redis

**Il modulo non emette nessun `*core.ApplicationError`**, quindi non ha codici né `Ambit`:
`redis.NewService` ritorna il `*goredis.Client` e gli errori risalgono così come li produce il
client (incluso `redis.Nil` per una chiave assente). Chi li espone via HTTP li avvolge nel
proprio `core.TechnicalError().WithCause(err)`, ottenendo `TECH500` con l'ambit dell'app.

È una scelta, non una dimenticanza: il modulo non ha operazioni proprie da nominare — è
l'accesso al client — e un codice per ogni comando Redis sarebbe un secondo vocabolario da
mantenere allineato a quello di go-redis.

## Errori sentinella del locker

`locker/` implementa `lock.Locker` su redsync (Redlock) e ritorna i due errori neutri di
`go-core-app/lock`:

| Errore | Origine | Significato |
|---|---|---|
| `lock.ErrNotAcquired` | `locker/locker.go:59` | il lock è già tenuto da un'altra replica. È **contesa, non un guasto**: lo scheduler (via `scheduler/gocronlock`) salta il tick. La correttezza è garantita dal claiming sul DB |
| `lock.ErrLockLost` | `locker/locker.go:92,97` | `Extend` fallito o `false`: il TTL redsync (~30s) è scaduto e un'altra replica può aver preso il lock. Il lavoro in corso non è più protetto |

Sono avvolti con `fmt.Errorf("redis lock extend %q: %w", ...)` — che nomina anche il backend:
vanno confrontati con `errors.Is`, mai con `==`.

## Errori di configurazione

Config non valida (tag `validate:` sulla `redis.Config`) o connessione impossibile all'avvio →
l'errore risale a fx e **l'app non parte**. Non c'è codice: non c'è chiamante da informare.
