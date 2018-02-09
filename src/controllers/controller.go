package controllers

import (
	"net/http"

	"github.com/SemyonL95/social-tournament-service/src/database"
)

func Run(db *database.DB) {
	router(db)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
