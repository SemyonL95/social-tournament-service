package main

import (
	"github.com/SemyonL95/social-tournament-service/src/database"
	"github.com/SemyonL95/social-tournament-service/src/controllers"
)

func main() {
	db, err := database.InitDatabaseConn()
	if err != nil {
		panic(err.Error())
	}

	controllers.Run(db)
}
