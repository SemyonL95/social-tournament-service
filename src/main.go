package main

import (
	"fmt"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func main() {
	var err error

	db, err = setDatabaseConn()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Database connections setup successfuly")

	serve()
}

func setDatabaseConn() (*sqlx.DB, error) {
	connInfo := fmt.Sprintf(
		"user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		"postgres",
		"postgres",
		"mypass",
		"db",
		"5432",
	)

	var err error

	db, err := sqlx.Open("postgres", connInfo)

	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
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
