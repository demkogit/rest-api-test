package store

type Config struct {
	DatabaseURL    string `toml:"database_url"`
	DatabaseName   string `toml:"database_name"`
	CollectionName string `toml:"collection_name"`
}

func NewConfig() *Config {
	return &Config{}
}
