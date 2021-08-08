package main

import "time"

type Conf struct {
	Port               int           `env:"GUARD_PORT" envDefault:"8000"`
	DSN                string        `env:"GUARD_DSN,required"`
	SecretKey          string        `env:"GUARD_SECRET_KEY,required"`
	AccessTTL          time.Duration `env:"GUARD_ACCESS_TTL" envDefault:"300s"`
	RefreshTTL         time.Duration `env:"GUARD_REFRESH_TTL" envDefault:"86400s"`
	CodeTTL            time.Duration `env:"GUARD_CODE_TTL" envDefault:"3600s"`
	BaseURL            string        `env:"GUARD_BASE_URL" envDefault:"http://localhost:8000"`
	GoogleClientID     string        `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string        `env:"GOOGLE_CLIENT_SECRET"`
}
