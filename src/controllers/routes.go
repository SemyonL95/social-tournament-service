package controllers

import (
	"net/http"

	"github.com/SemyonL95/social-tournament-service/src/database"
)

func router(db *database.DB) {
	http.HandleFunc("/fund", func(w http.ResponseWriter, r *http.Request) {
		fund(db, w, r)
	})
	http.HandleFunc("/take", func(w http.ResponseWriter, r *http.Request) {
		take(db, w, r)
	})
	http.HandleFunc("/announceTournament", func(w http.ResponseWriter, r *http.Request) {
		announceTournament(db, w, r)
	})
	http.HandleFunc("/joinTournament", func(w http.ResponseWriter, r *http.Request) {
		joinTournament(db, w, r)
	})
	http.HandleFunc("/resultTournament", func(w http.ResponseWriter, r *http.Request) {
		resultTournament(db, w, r)
	})
	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		reset(db, w, r)
	})
}
