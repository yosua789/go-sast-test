package api

import (
	"assist-tix/config"

	sentry "github.com/getsentry/sentry-go"
)

func InitSentry(env *config.EnvironmentVariable) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:           env.Sentry.Dsn,
		EnableTracing: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for tracing.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	})
}
