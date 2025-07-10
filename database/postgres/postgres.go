package postgres

import (
	"assist-tix/config"
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type WrapDatabase struct {
	Conn *pgxpool.Pool
}

const MIGRATION_LOCATIONS = "database/migrations"

func NewDBConnection(env *config.EnvironmentVariable) *pgxpool.Pool {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s",
		env.Database.Postgres.User,
		env.Database.Postgres.Password,
		env.Database.Postgres.Host,
		env.Database.Postgres.Name,
	)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal().Err(err).Str("database", env.Database.Postgres.Name).Msg("[x] failed to parse connection config for postgres")
		panic(err)
	}

	conn, err := pgxpool.New(context.Background(), config.ConnString())
	if err != nil {
		log.Fatal().Err(err).Str("database", env.Database.Postgres.Name).Msg("[x] failed to connect to postgres")
		panic(err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		log.Fatal().Err(err).Str("database", env.Database.Postgres.Name).Msg("[x] failed to ping postgres")
		panic(err)
	}

	return conn
}

func InitMigrations(env *config.EnvironmentVariable) error {
	log.Info().Msg("Checking migrations")

	fmt.Println(env.GetDBUrl())

	m, err := migrate.New(
		fmt.Sprintf("file://%s", MIGRATION_LOCATIONS),
		env.GetDBUrl(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed load migrations")
		return err
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("Database has not changed")
			return nil
		}
		log.Fatal().Err(err).Msg("Failed to run migration")
		return err
	}

	return nil
}
