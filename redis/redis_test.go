package redis

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func config(t *testing.T, mr *miniredis.Miniredis) *Config {
	t.Helper()
	host, port, _ := strings.Cut(mr.Addr(), ":")
	p, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}
	return &Config{Address: host, Port: p}
}

// Il client fornito è utilizzabile e si chiude all'arresto.
func TestNewService_Utilizzabile(t *testing.T) {
	mr := miniredis.RunT(t)
	var c *goredis.Client
	app := fxtest.New(t, fx.Supply(config(t, mr)), fx.Provide(NewService), fx.Populate(&c))
	app.RequireStart()
	if err := c.Set(context.Background(), "k", "v", 0).Err(); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got, _ := mr.Get("k"); got != "v" {
		t.Fatalf("valore su redis = %q", got)
	}
	app.RequireStop()
	if err := c.Ping(context.Background()).Err(); err == nil {
		t.Fatal("il client deve essere chiuso dopo lo stop")
	}
}

// Un Redis irraggiungibile o una password sbagliata fermano l'avvio: prima l'errore emergeva alla
// prima operazione.
func TestNewService_FailFastAllAvvio(t *testing.T) {
	mr := miniredis.RunT(t)
	mr.RequireAuth("giusta")
	cfg := config(t, mr)
	cfg.Password = "sbagliata"
	app := fx.New(fx.NopLogger, fx.Supply(cfg), fx.Provide(NewService), fx.Invoke(func(*goredis.Client) {}))
	if err := app.Start(context.Background()); err == nil {
		t.Fatal("avvio riuscito con una password sbagliata")
	}
}
