package config

import "os"

type Config struct {
	HTTPAddr         string
	Driver           string
	DSN              string
	MercagoPagoToken string
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	return Config{
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		Driver:           getEnv("DB_DRIVER", "postgres"),
		DSN:              getEnv("DSN", "host=localhost user=postgres password=postgres dbname=powermixdb port=5432 sslmode=disable"),
		MercagoPagoToken: getEnv("MERCAGO_PAGO_TOKEN", "TEST"),
	}
}
