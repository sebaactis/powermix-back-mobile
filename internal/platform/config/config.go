package config

import (
	"fmt"
	"os"
	"strings"
)

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

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:         os.Getenv("HTTP_ADDR"),
		Driver:           os.Getenv("DB_DRIVER"),
		DSN:              os.Getenv("DSN"),
		MercagoPagoToken: os.Getenv("MERCAGO_PAGO_TOKEN"),
		CoffejiKey:       os.Getenv("COFFEJI_KEY"),
		CoffejiSecret:    os.Getenv("COFFEJI_SECRET"),
		ResendKey:        os.Getenv("RESEND_API_KEY"),
		HashToken:        os.Getenv("JWT_REFRESH_HASH"),
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	required := map[string]string{
		"HTTP_ADDR":          c.HTTPAddr,
		"DB_DRIVER":          c.Driver,
		"DSN":                c.DSN,
		"MERCAGO_PAGO_TOKEN": c.MercagoPagoToken,
		"COFFEJI_KEY":        c.CoffejiKey,
		"COFFEJI_SECRET":     c.CoffejiSecret,
		"RESEND_API_KEY":     c.ResendKey,
		"JWT_REFRESH_HASH":   c.HashToken,
	}
	for key, val := range required {
		if strings.TrimSpace(val) == "" {
			return fmt.Errorf("variable de entorno requerida no configurada: %s", key)
		}
	}
	return nil
}
