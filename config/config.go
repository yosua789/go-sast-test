package config

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func LoadEnv() (env *EnvironmentVariable, err error) {
	log.Info().Msg("Load Env Here")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Msg("viper error read config")
	}

	err = viper.Unmarshal(&env)
	if err != nil {
		log.Error().Err(err).Msg("viper error unmarshal config")
	}
	return
}

func SetDefaultConfig(v *viper.Viper) {
	v.SetDefault("DATABASE.TIMEOUT.PING", "5s")
	v.SetDefault("DATABASE.TIMEOUT.READ", "5s")
	v.SetDefault("DATABASE.TIMEOUT.WRITE", "5s")
}

type EnvironmentVariable struct {
	App struct {
		Host  string `mapstucture:"HOST"`
		Port  int    `mapstructure:"PORT"`
		Mode  string `mapstructure:"MODE"`
		Debug bool   `mapstructure:"DEBUG"`
	} `mapstructure:"APP"`
	Api struct {
		CorsEnable bool   `mapstructure:"CORS_ENABLE"`
		BasePath   string `mapstructure:"BASE_PATH"`
	} `mapstructure:"API"`
	Database struct {
		Postgres struct {
			Scheme            string `mapstructure:"SCHEME"`
			Host              string `mapstructure:"HOST"`
			Port              string `mapstructure:"PORT"`
			User              string `mapstructure:"USER"`
			Password          string `mapstructure:"PASSWORD"`
			Name              string `mapstructure:"NAME"`
			MigrationLocation string `mapstructure:"MIGRATION_LOCATION"`
		} `mapstructure:"POSTGRES"`
		Timeout struct {
			Ping  time.Duration `mapstructure:"PING"`
			Read  time.Duration `mapstructure:"READ"`
			Write time.Duration `mapstructure:"Write"`
		} `mapstructure:"TIMEOUT"`
	} `mapstructure:"DATABASE"`
	Nats struct {
		Addr   string `mapstructure:"ADDR"`
		Port   int    `mapstructure:"PORT"`
		Token  string `mapstructure:"TOKEN"`
		Stream struct {
			Mailer string `mapstructutre:"MAILER"`
		} `mapstructure:"STREAM"`
	} `mapstructure:"NATS"`
	Mailer struct {
		AssetsPath string `mapstructure:"ASSETS_PATH"`
	} `mapstructure:"MAILER"`
	Swagger struct {
		Host     string `mapstructure:"HOST"`
		Protocol string `mapstructure:"PROTOCOL"`
	} `mapstructure:"SWAGGER"`
	FileUpload struct {
		MaxSize int `mapstructure:"MAX_SIZE"`
	} `mapstructure:"FILE_UPLOAD"`
}

func (e *EnvironmentVariable) GetDBDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", e.Database.Postgres.Host, e.Database.Postgres.Port, e.Database.Postgres.User, e.Database.Postgres.Password, e.Database.Postgres.Name)
}

func (e *EnvironmentVariable) GetDBUrl() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", e.Database.Postgres.Scheme, e.Database.Postgres.User, e.Database.Postgres.Password, e.Database.Postgres.Host, e.Database.Postgres.Port, e.Database.Postgres.Name)
}
