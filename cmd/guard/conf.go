package main

import "time"

type Conf struct {
	Debug                 bool          `env:"GUARD_DEBUG" envDefault:"false"`
	Port                  int           `env:"GUARD_PORT" envDefault:"8000"`
	LogLevel              string        `env:"GUARD_LOG_LEVEL" envDefault:"info"`
	DBDriver              string        `env:"GUARD_DBDRIVER" envDefault:"sqlite"`
	DBMaxIddleConn        int           `env:"GUARD_DB_MAX_IDDLE_CONN" envDefault:"16"`
	DBMaxOpenConn         int           `env:"GUARD_DB_MAX_OPEN_CONN" envDefault:"128"`
	DBConnMaxLifetime     time.Duration `env:"GUARD_DB_CONN_MAX_LIFETIME" envDefault:"3600s"`
	DBConnMaxIddleTime    time.Duration `env:"GUARD_DB_CONN_MAX_IDDLE_TIME" envDefault:"300s"`
	DSN                   string        `env:"GUARD_DSN,required"`
	SecretKey             string        `env:"GUARD_SECRET_KEY,required"`
	AccessTTL             time.Duration `env:"GUARD_ACCESS_TTL" envDefault:"300s"`
	RefreshTTL            time.Duration `env:"GUARD_REFRESH_TTL" envDefault:"86400s"`
	CodeTTL               time.Duration `env:"GUARD_CODE_TTL" envDefault:"3600s"`
	BaseURL               string        `env:"GUARD_BASE_URL" envDefault:"http://localhost:8000"`
	ClaimsEndpoint        string        `env:"GUARD_CLAIMS_ENDPOINT,required"`
	ClaimsAuthHeader      string        `env:"GUARD_CLAIMS_AUTH_HEADER,required"`
	ClaimsAuthToken       string        `env:"GUARD_CLAIMS_AUTH_TOKEN,required"`
	ClaimsJsonPath        string        `env:"GUARD_CLAIMS_JSON_PATH,required"`
	HttpProfileEndppint   string        `env:"GUARD_HTTP_PROFILE_ENDPOINT,required"`
	HttpProfileAuthHeader string        `env:"GUARD_HTTP_PROFILE_AUTH_HEADER,required"`
	HttpProfileAuthToken  string        `env:"GUARD_HTTP_PROFILE_AUTH_TOKEN,required"`
	AppleClientID         string        `env:"APPLE_CLIENT_ID"`
	AppleClientSecret     string        `env:"APPLE_CLIENT_SECRET"`
	GoogleClientID        string        `env:"GOOGLE_CLIENT_ID"`
	GoogleSecret          string        `env:"GOOGLE_CLIENT_SECRET"`
	FacebookClientID      string        `env:"FACEBOOK_CLIENT_ID"`
	FacebookSecret        string        `env:"FACEBOOK_CLIENT_SECRET"`
	TwitterClientID       string        `env:"TWITTER_CLIENT_ID"`
	TwitterSecret         string        `env:"TWITTER_CLIENT_SECRET"`
	VkClientID            string        `env:"VK_CLIENT_ID"`
	VkSecret              string        `env:"VK_CLIENT_SECRET"`
	YandexClientID        string        `env:"YANDEX_CLIENT_ID"`
	YandexSecret          string        `env:"YANDEX_CLIENT_SECRET"`
}
