package env

import "github.com/caarlos0/env/v11"

type StaticEnvStruct struct {
	ServerPort int `env:"SERVER_PORT"`

	DbHost     string `env:"DB_HOST"`
	DbPort     int    `env:"DB_PORT"`
	DbUser     string `env:"DB_USER"`
	DbPassword string `env:"DB_PASSWORD"`
	DbName     string `env:"DB_NAME"`
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
