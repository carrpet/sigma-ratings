package main

type Config struct {
	Database struct {
		User     string `env:"POSTGRES_USER,required"`
		DBName   string `env:"POSTGRES_DB,required"`
		Password string `env:"POSTGRES_PASSWORD,required"`
	}

	SanctionsBackend struct {
		URL string `env:"SANCTIONS_URL,default=http://sigmaratings.s3.us-east-2.amazonaws.com/eu_sanctions.csv"`
	}

	FrontEnd struct {
		Port string `env:"SRV_PORT,default=8080"`
	}
}
