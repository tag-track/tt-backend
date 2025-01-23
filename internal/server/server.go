package server

import (
	"Backend/internal/database"
	"Backend/internal/env"
	"Backend/internal/objectstore"
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
		log.Fatalln(err)
	}

	if err := db.Migrate(context.Background()); err != nil {
		log.Fatalln(err)
	}

	return db
}

func Serve() {

	e := env.GetStaticEnv()
	db := createDbInstance()
	objStore := objectstore.NewMinioAdapter()

	mainRouter := http.NewServeMux()

	mainRouter.
		Handle(
			"/api/v1/",
			http.StripPrefix(
				"/api/v1",
				middleware.Apply(
					v1.Router(),
					middleware.ApplyTimeout(2000*time.Millisecond),
					middleware.ApplyAttachObjStore(objStore),
					middleware.ApplyAttachDb(db),
				),
			),
		)

	loggedRouter := middleware.LoggingMiddleware(mainRouter)

	log.Printf("Starting server on 0.0.0.0:%v", e.ServerPort)
	_ = http.ListenAndServe(
		fmt.Sprintf("0.0.0.0:%v", e.ServerPort),
		loggedRouter,
	)
}
