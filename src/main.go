package main

import (
	"net/http"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/SemyonL95/social-tournament-service/src/database"
	"github.com/SemyonL95/social-tournament-service/src/validators"
	"database/sql"
)

func main() {
	db, err := database.InitDatabaseConn()
	if err != nil {
		panic(err.Error())
	}

	serve(db)
}

func serve(db *database.DB) {
	http.HandleFunc("/fund", func(w http.ResponseWriter, r *http.Request) {
		fund(db, w, r)
	})

	http.HandleFunc("/take", func(w http.ResponseWriter, r *http.Request) {
		take(db, w, r)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

//TODO REFACTOR ALL THIS CRAP
func fund(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	//TODO refactor from credits to points everywhere
	username := r.FormValue("playerId")
	credits := r.FormValue("points")

	validated := validators.ValidateUsername(username)
	if !validated {
		http.Error(w, "playerId is required and have to be a string A-Za-z0-9 min: 1, max: 20 characters \n", http.StatusUnprocessableEntity)
		return
	}

	parsedCredits, err := strconv.ParseFloat(credits, 64)
	if err != nil {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	validated = validators.ValidateFloatNotNegtive(parsedCredits)
	if !validated {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	user, err := db.FundOrCreateUser(username, parsedCredits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("user " + user.Username + " funded"))
	return
}

func take(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("playerId")
	credits := r.FormValue("points")

	validated := validators.ValidateUsername(username)
	if !validated {
		http.Error(w, "playerId is required and have to be a string A-Za-z0-9 min: 1, max: 20 characters \n", http.StatusUnprocessableEntity)
		return
	}

	parsedCredits, err := strconv.ParseFloat(credits, 64)
	if err != nil {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	validated = validators.ValidateFloatNotNegtive(parsedCredits)
	if !validated {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	user, err := db.TakePointsFromUser(username, parsedCredits)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			http.Error(w, err.Error(), http.StatusNotFound)
			break
		case database.ErrNotEnoughMoney:
			http.Error(w, err.Error(), http.StatusForbidden)
			break
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Write([]byte("User " + user.Username + " : points " + strconv.FormatFloat(user.Credits, 'f', 2, 64)))
	return
}
