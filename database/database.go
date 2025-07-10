package database

import (
	"assist-tix/config"
	"assist-tix/database/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WrapDB struct {
	Postgres *pgxpool.Pool
}

func InitDB(env *config.EnvironmentVariable) *WrapDB {
	postgresDB := postgres.NewDBConnection(env)

	// Init migrations
	err := postgres.InitMigrations(env)
	if err != nil {
		return nil
	}

	return &WrapDB{
		Postgres: postgresDB,
	}
}
