package env

import "github.com/caarlos0/env/v11"

type StaticEnvStruct struct {
	ServerPort int `env:"SERVER_PORT"`

	DbHost        string `env:"DB_HOST"`
	DbPort        int    `env:"DB_PORT"`
	DbUser        string `env:"DB_USER"`
	DbPassword    string `env:"DB_PASSWORD"`
	DbName        string `env:"DB_NAME"`
	MinioHost     string `env:"MINIO_HOST"`
	MinioPort     string `env:"MINIO_PORT_API"`
	MinioUser     string `env:"MINIO_USER"`
	MinioPassword string `env:"MINIO_PASSWORD"`
	MinioBucket   string `env:"MINIO_BUCKET"`
}

var (
	staticEnv *StaticEnvStruct
)

func LoadStaticEnv() {
	staticEnv = &StaticEnvStruct{}
	if err := env.Parse(staticEnv); err != nil {
		panic(err)
	}
}

func GetStaticEnv() *StaticEnvStruct {
	if staticEnv == nil {
		LoadStaticEnv()
	}

	return staticEnv
}
