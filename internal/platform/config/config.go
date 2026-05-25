package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr               string
	Driver                 string
	DSN                    string
	MercagoPagoToken       string
	CoffejiKey             string
	CoffejiSecret          string
	ResendKey              string
	HashToken              string

	// ProdeEnabled activa/desactiva toda la feature PRODE.
	// false = las rutas /api/v1/prode/* no se registran, las tablas existen pero
	// no se usan. Sirve como kill switch para rollback sin perder datos.
	ProdeEnabled           bool
	// ProdeMaintenanceEnabled activa la proteccion de endpoints admin via
	// X-Prode-Admin-Key. Si es true, PRODE_ADMIN_API_KEY es obligatoria.
	ProdeMaintenanceEnabled bool
	ProdeAdminAPIKey       string
	ProdeAdminEmails       []string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:                os.Getenv("HTTP_ADDR"),
		Driver:                  os.Getenv("DB_DRIVER"),
		DSN:                     os.Getenv("DSN"),
		MercagoPagoToken:        os.Getenv("MERCAGO_PAGO_TOKEN"),
		CoffejiKey:              os.Getenv("COFFEJI_KEY"),
		CoffejiSecret:           os.Getenv("COFFEJI_SECRET"),
		ResendKey:               os.Getenv("RESEND_API_KEY"),
		HashToken:               os.Getenv("JWT_REFRESH_HASH"),
		ProdeEnabled:            os.Getenv("PRODE_ENABLED") == "true",
		ProdeMaintenanceEnabled: os.Getenv("PRODE_MAINTENANCE_ENABLED") == "true",
		ProdeAdminAPIKey:        os.Getenv("PRODE_ADMIN_API_KEY"),
	}

	if emails := os.Getenv("PRODE_ADMIN_EMAILS"); emails != "" {
		parts := strings.Split(emails, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		cfg.ProdeAdminEmails = parts
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// IsProdeEnabled indica si la funcionalidad PRODE está habilitada.
func (c Config) IsProdeEnabled() bool {
	return c.ProdeEnabled
}

// IsMaintenanceEnabled indica si el modo mantenimiento PRODE está activo.
func (c Config) IsMaintenanceEnabled() bool {
	return c.ProdeMaintenanceEnabled
}

// AdminAPIKey devuelve la clave de administración PRODE configurada.
func (c Config) AdminAPIKey() string {
	return c.ProdeAdminAPIKey
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

	if c.ProdeMaintenanceEnabled && strings.TrimSpace(c.ProdeAdminAPIKey) == "" {
		return fmt.Errorf("PRODE_ADMIN_API_KEY is required when PRODE_MAINTENANCE_ENABLED is true")
	}

	return nil
}
