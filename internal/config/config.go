package config

import "os"

type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
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
