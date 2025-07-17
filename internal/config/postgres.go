package config

func PostgresHost() string {
	return cfg.Database.Postgres.Host
}

func PostgresUsername() string {
	return cfg.Database.Postgres.Username
}

func PostgresPassword() string {
	return cfg.Database.Postgres.Password
}

func PostgresDatabase() string {
	return cfg.Database.Postgres.Database
}

func PostgresPort() string {
	return cfg.Database.Postgres.Port
}

func PostgresMigrationsPath() string {
	return cfg.Database.Postgres.MigrationsPath
}
