package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

func main() {
	connInfo := fmt.Sprintf(
		"user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		"postgres",
		"postgres",
		"mypass",
		"db",
		"5432",
	)

	var err error
	db, err = sql.Open("postgres", connInfo)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(
		`create table if not exists mydata (
			id serial primary key,
			val integer not null
		)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", serveIndex)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func serveIndex(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Hello, World!\n")

	_, err := db.Exec("insert into mydata(val) values(1)")

	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("select id from mydata")
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var id int

		err = rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintf(w, "ID: %d\n", id)
	}
}
