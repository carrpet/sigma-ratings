package main

type Config struct {
	Database struct {
		User     string `yaml:"user", envconfig: "POSTGRES_USER"`
		DBName   string `yaml:"dbName", envconfig: "POSTGRES_DB"`
		Password string `yaml:"password", envconfig: "POSTGRES_PASSWORD"`
	} `yaml:"database"`

	SanctionsBackend struct {
		URL string `yaml:"url"`
	} `yaml:"sanctionsBackend"`

	FrontEnd struct {
		Port string `yaml:"port"`
	}
}
