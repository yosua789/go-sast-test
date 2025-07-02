package postgres

import (
	"assist-tix/config"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type WrapDatabase struct {
	Conn *pgxpool.Pool
}

func InitDatabase(env *config.EnvironmentVariable) *WrapDatabase {
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

	return &WrapDatabase{
		Conn: conn,
	}
}
