package config

import "os"

type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	Auth   AuthConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type MySQLConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

type AuthConfig struct {
	TokenSecret        string
	TokenExpireSeconds int64
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Host: env("SERVER_HOST", "0.0.0.0"),
			Port: env("SERVER_PORT", "8080"),
		},
		MySQL: MySQLConfig{
			Host:     env("MYSQL_HOST", "127.0.0.1"),
			Port:     env("MYSQL_PORT", "3306"),
			Database: env("MYSQL_DATABASE", "jiaozi"),
			Username: env("MYSQL_USERNAME", "root"),
			Password: env("MYSQL_PASSWORD", ""),
		},
		Auth: AuthConfig{
			TokenSecret:        env("AUTH_TOKEN_SECRET", "dev-secret"),
			TokenExpireSeconds: 604800,
		},
	}
}

func (c Config) ServerAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
