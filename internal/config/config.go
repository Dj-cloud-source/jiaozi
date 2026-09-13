package config

import "os"

type Config struct {
	Host string
	Port string
}

func Load() Config {
	return Config{
		Host: env("SERVER_HOST", "0.0.0.0"),
		Port: env("SERVER_PORT", "8080"),
	}
}

func (c Config) ServerAddress() string {
	return c.Host + ":" + c.Port
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
