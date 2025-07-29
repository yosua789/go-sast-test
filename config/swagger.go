package config

import (
	"assist-tix/docs"
	"path"
)

const (
	ReleaseVersion = "1.2.3"
)

func InitSwagger(env *EnvironmentVariable) {
	if env.App.Debug {
		docs.SwaggerInfo.Version = ReleaseVersion
		docs.SwaggerInfo.Host = env.Swagger.Host
		docs.SwaggerInfo.BasePath = path.Join(env.Api.BasePath, "/api/v1/")
		docs.SwaggerInfo.Schemes = []string{"http", "https"}
	}
}
