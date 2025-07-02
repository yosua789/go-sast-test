package database

import (
	"assist-tix/config"
	"assist-tix/database/postgres"
)

type WrapDB struct {
	Postgres *postgres.WrapDatabase
}

func InitDB(env *config.EnvironmentVariable) *WrapDB {
	postgresDB := postgres.InitDatabase(env)

	return &WrapDB{
		Postgres: postgresDB,
	}
}
