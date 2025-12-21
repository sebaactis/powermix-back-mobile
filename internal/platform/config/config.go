package config

import "os"

type Config struct {
	HTTPAddr         string
	Driver           string
	DSN              string
	MercagoPagoToken string
	CoffejiKey       string
	CoffejiSecret    string
	ResendKey        string
	HashToken        string
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
		DSN:              getEnv("DSN", "postgresql://postgres:Powermix0403@@db.kzlzslphwkmypcblgntn.supabase.co:5432/postgres"),
		MercagoPagoToken: getEnv("MERCAGO_PAGO_TOKEN", "TEST"),
		CoffejiKey:       getEnv("COFFEJI_KEY", "TEST"),
		CoffejiSecret:    getEnv("COFFEJI_SECRET", "TEST"),
		ResendKey:        getEnv("RESEND_API_KEY", "TEST"),
		HashToken:        getEnv("JWT_REFRESH_HASH", "TEST"),
	}
}
