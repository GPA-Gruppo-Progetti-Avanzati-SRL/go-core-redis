package redis

// Config configures a general-purpose Redis client. It is intentionally minimal
// (address, port, password); extend as needed for TLS / DB index / pool sizing.
type Config struct {
	Enable   bool   `yaml:"enable" mapstructure:"enable" json:"enable"`
	Address  string `yaml:"address" mapstructure:"address" json:"address"`
	Port     int    `yaml:"port" mapstructure:"port" json:"port"`
	Password string `yaml:"password" mapstructure:"password" json:"password"`
}
