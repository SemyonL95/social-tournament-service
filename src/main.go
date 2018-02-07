package main

import (
	"net/http"

	_ "github.com/lib/pq"

	"./models"
	"./db"
	"fmt"
)


func main() {
	conn, err := db.InitDatabaseConn()
	if err != nil {
		panic(err.Error())
	}
	model := models.InitModel(conn)

	model.Testdb()

	serve()
}


func serve() {
	http.HandleFunc("/", indexPage)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Hello world"))
}
