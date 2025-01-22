package server

import (
	"Backend/internal/database"
	"Backend/internal/env"
	"Backend/internal/server/handler/api/v1"
	"Backend/internal/server/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func createDbInstance() *database.GormPgAdapter {

	e := env.GetStaticEnv()

	db, err := database.CreateGormPgAdapter(
		e.DbHost,
		e.DbUser,
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
			"/api/v1/",
			http.StripPrefix(
				"/api/v1",
				middleware.Apply(
					v1.Router(),
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
