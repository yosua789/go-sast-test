package config

import (
	"assist-tix/dto"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func LoadEnv() (env *EnvironmentVariable, err error) {
	log.Info().Msg("Load Env Here")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	SetDefaultConfig(viper.GetViper())

	err = viper.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Msg("viper error read config")
	}

	err = viper.Unmarshal(&env)
	if err != nil {
		log.Error().Err(err).Msg("viper error unmarshal config")
	}

	fmt.Println(env.Database.Timeout)

	// Check credential is filename
	var credential dto.GCPServiceAccount
	err = json.Unmarshal([]byte(env.Storage.GCS.Credential), &credential)
	if err != nil {
		log.Warn().Msg("failed convert object GCS credential, checking by filename")
		_, err = os.Stat(env.Storage.GCS.Credential)
		if err == nil || !os.IsNotExist(err) {
			file, err := os.Open(env.Storage.GCS.Credential)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			bytes, err := io.ReadAll(file)
			if err != nil {
				log.Error().Err(err).Msg("failed to read file credential")
				return nil, err
			}

			errJson := json.Unmarshal(bytes, &credential)
			if errJson != nil {
				log.Warn().Msg("failed convert object GCS credential by filename")
				return nil, err
			}

			err = nil
		}
	}

	env.Redis.Host = fmt.Sprintf("%s:%s", env.Redis.Address, env.Redis.Port)
	env.Storage.GCS.CredentialObj = credential

	return
}

func SetDefaultConfig(v *viper.Viper) {
	v.SetDefault("DATABASE.TIMEOUT.PING", "1s")
	v.SetDefault("DATABASE.TIMEOUT.READ", "5s")
	v.SetDefault("DATABASE.TIMEOUT.WRITE", "5s")

	v.SetDefault("ASYNQ.DEADLINE_DURATION", "30s")
}

type EnvironmentVariable struct {
	App struct {
		Host  string `mapstucture:"HOST"`
		Port  int    `mapstructure:"PORT"`
		Mode  string `mapstructure:"MODE"`
		Debug bool   `mapstructure:"DEBUG"`

		AutoAssignSeat bool `mapstructure:"AUTO_ASSIGN_SEAT"` // It will disable validation seat
	} `mapstructure:"APP"`
	Api struct {
		CorsEnable bool   `mapstructure:"CORS_ENABLE"`
		BasePath   string `mapstructure:"BASE_PATH"`
		Url        string `mapstructure:"URL"`
	} `mapstructure:"API"`
	AccessToken struct {
		SecretKey string `mapstructure:"SECRET_KEY"`
	} `mapstructure:"ACCESS_TOKEN"`
	Redis struct {
		Address  string `mapstructure:"ADDRESS"`
		Port     string `mapstructure:"PORT"`
		Username string `mapstructure:"USERNAME"`
		Password string `mapstructure:"PASSWORD"`
		Host     string `mapstructure:"HOST"`
	} `mapstructure:"REDIS"`
	Database struct {
		Postgres struct {
			UseMigration   bool   `mapstructure:"USE_MIGRATION"`
			Scheme         string `mapstructure:"SCHEME"`
			Host           string `mapstructure:"HOST"`
			Port           string `mapstructure:"PORT"`
			User           string `mapstructure:"USER"`
			Password       string `mapstructure:"PASSWORD"`
			Name           string `mapstructure:"NAME"`
			MaxConnections int    `mapstructure:"MAX_CONNECTIONS"`
			MaxIdleTime    int    `mapstructure:"MAX_IDLE_TIME"`
		} `mapstructure:"POSTGRES"`
		Timeout struct {
			Ping  time.Duration `mapstructure:"PING"`
			Read  time.Duration `mapstructure:"READ"`
			Write time.Duration `mapstructure:"WRITE"`
		} `mapstructure:"TIMEOUT"`
	} `mapstructure:"DATABASE"`
	Nats struct {
		Addr     string `mapstructure:"ADDR"`
		Port     string `mapstructure:"PORT"`
		Token    string `mapstructure:"TOKEN"`
		Subjects struct {
			SendBill    string `mapstructure:"SEND_BILL"`
			SendInvoice string `mapstructure:"SEND_INVOICE"`
			SendETicket string `mapstructure:"SEND_ETICKET"`
		} `mapstructure:"SUBJECTS"`
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

	Transaction struct {
		ExpirationDuration time.Duration `mapstructure:"EXPIRATION_DURATION"`
	} `mapstructure:"TRANSACTION"`
	Paylabs struct {
		BaseUrl         string `mapstructure:"BASE_URL"`
		AccountID       string `mapstructure:"ACCOUNT_ID"`
		PublicKey       string `mapstructure:"PUBLIC_KEY"`       // paylabs public key in PEM format
		PrivateKey      string `mapstructure:"PRIVATE_KEY"`      // our private key in PEM format
		PaymentDuration int    `mapstructure:"PAYMENT_DURATION"` // Duration in seconds
		ActivePayment   bool   `mapstructure:"ACTIVE_PAYMENT"`   // Enable or disable payment
	} `mapstructure:"PAYLABS"`
	Storage struct {
		Type string `mapstructure:"TYPE"`
		GCS  struct {
			BucketName          string `mapstructure:"BUCKET_NAME"`
			Credential          string `mapstructure:"CREDENTIAL"`
			CredentialObj       dto.GCPServiceAccount
			SignedUrlExpiration time.Duration `mapstructure:"SIGNED_URL_EXPIRATION"`
		} `mapstructure:"GCS"`
	} `mapstructure:"STORAGE"`
	GarudaID struct {
		BaseUrl string `mapstructure:"BASE_URL"`
		ApiKey  string `mapstructure:"API_KEY"`
		// IsMock?  bool   `mapstructure:"IS_MOCK"`
		MinimumAge int `mapstructure:"MINIMUM_AGE"` // Minimum age in years for Garuda ID verification
	} `mapstructure:"GARUDA_ID"`
	Sentry struct {
		Dsn string `mapstructure:"DSN"`
	} `mapstructure:"SENTRY"`
	Asynq struct {
		ProcessTimeout time.Duration `mapstructure:"PROCESS_TIMEOUT"`
	} `mapstructure:"ASYNQ"`
}

func (e *EnvironmentVariable) GetDBDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", e.Database.Postgres.Host, e.Database.Postgres.Port, e.Database.Postgres.User, e.Database.Postgres.Password, e.Database.Postgres.Name)
}

func (e *EnvironmentVariable) GetDBUrl() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", e.Database.Postgres.Scheme, e.Database.Postgres.User, e.Database.Postgres.Password, e.Database.Postgres.Host, e.Database.Postgres.Port, e.Database.Postgres.Name)
}
