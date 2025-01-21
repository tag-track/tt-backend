package env

import "github.com/caarlos0/env/v11"

type StaticEnvStruct struct {
	ServerPort int

	DbHost     string
	DbPort     int
	DbUser     string
	DbPassword string
	DbName     string
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
