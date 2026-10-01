package redis

// Config configures a general-purpose Redis client. It is intentionally minimal
// (address, port, password); extend as needed for TLS / DB index / pool sizing.
//
// Non ha un campo `enable`: l'attivazione la decidono i modes passati a Module, come per gli altri
// driver. C'era, e nessuno lo leggeva — `enable: false` nello YAML non spegneva nulla.
type Config struct {
	Address  string `yaml:"address" mapstructure:"address" json:"address"`
	Port     int    `yaml:"port" mapstructure:"port" json:"port"`
	Password string `yaml:"password" mapstructure:"password" json:"password"`
}
