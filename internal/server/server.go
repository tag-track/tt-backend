package server

import (
	"Backend/internal/database"
	"Backend/internal/env"
	"Backend/internal/server/handler/api"
	"Backend/internal/server/middleware"
	"fmt"
	"net/http"
	"time"
)

func createDbInstance() *database.GormPgAdapter {

	e := env.GetStaticEnv()

	db, err := database.CreateGormPgAdapter(
		e.DbHost,
		e.DbHost,
		e.DbPassword,
		e.DbPort,
		e.DbName,
	)
	if err != nil {
		panic(err)
	}

	return db
}

func Serve() {

	e := env.GetStaticEnv()
	db := createDbInstance()

	mainRouter := http.NewServeMux()

	mainRouter.
		Handle(
			"/api/",
			http.StripPrefix(
				"/api",
				middleware.Apply(
					api.Router(),
					middleware.ApplyTimeout(150*time.Millisecond),
					middleware.ApplyAttachDb(db),
				),
			),
		)

	_ = http.ListenAndServe(
		fmt.Sprintf("0.0.0.0:%v", e.DbPort),
		mainRouter,
	)
}
