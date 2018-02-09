package main

import (
	"net/http"
	"strconv"
	"database/sql"
	"log"
	"fmt"
	"regexp"
	"strings"

	"github.com/lib/pq"

	"github.com/SemyonL95/social-tournament-service/src/database"
	"github.com/SemyonL95/social-tournament-service/src/validators"
)

var BackersRegExp = regexp.MustCompile(`bakerId=[A-Za-z0-9]*`)
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
	http.HandleFunc("/announceTournament", func(w http.ResponseWriter, r *http.Request) {
		announceTournament(db, w, r)
	})
	http.HandleFunc("/joinTournament", func(w http.ResponseWriter, r *http.Request) {
		joinTournament(db, w, r)
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

	validated = validators.ValidateFloatNotNegative(parsedCredits)
	if !validated {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	err = db.FundOrCreateUser(username, parsedCredits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("user " + username + " funded"))
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

	validated = validators.ValidateFloatNotNegative(parsedCredits)
	if !validated {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	err = db.TakePointsFromUser(username, parsedCredits)
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

	w.Write([]byte("Points has been taken successfully"))
	return
}

func announceTournament(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	deposit := r.FormValue("deposit")

	if id == "" {
		log.Println("id empty")
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	parsedId, err := strconv.Atoi(id)
	if err != nil {
		log.Println("id convertion error")
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if parsedId <= 0 {
		log.Println("id <= 0")
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	parsedDeposit, err := strconv.ParseFloat(deposit, 64)
	if err != nil {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	validated := validators.ValidateFloatNotNegative(parsedDeposit)
	if !validated {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	err = db.CreateTournament(parsedId, parsedDeposit)
	if err != nil {
		var errMsg string
		if driverError, ok := err.(*pq.Error); ok && driverError.Code == "23505" {

			//See postgres errors codes https://www.postgresql.org/docs/10/static/errcodes-appendix.html (unique_violation)
			errMsg = fmt.Sprintf("tournament with id - %d already exists", parsedId)
			http.Error(w, errMsg, http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Tournament has been created successfully"))
	return
}

func joinTournament(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	rawBakersIds := BackersRegExp.FindAllString(r.URL.RawQuery, -1)
	tournamentId := r.FormValue("tournamentId")
	username := r.FormValue("playerId")


	var bakersIds []string
	for _,str:= range rawBakersIds {
		splitedStr := strings.Split(str, "=")
		validated := validators.ValidateUsername(splitedStr[1])
		if !validated {
			http.Error(w, "bakerId have to be a string A-Za-z0-9 min: 1, max: 20 characters", http.StatusUnprocessableEntity)
			return
		}
		bakersIds = append(bakersIds, splitedStr[1])
	}

	if tournamentId == "" {
		log.Println("id empty")
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	parsedTournamentId, err := strconv.Atoi(tournamentId)
	if err != nil {
		log.Println("id convertion error")
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if parsedTournamentId <= 0 {
		log.Println("id <= 0")
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	validated := validators.ValidateUsername(username)
	if !validated {
		http.Error(w, "playerId is required and have to be a string A-Za-z0-9 min: 1, max: 20 characters \n", http.StatusUnprocessableEntity)
		return
	}

	err = db.AssignToTournament(parsedTournamentId, username, bakersIds)
	if err != nil {
		//TODO handle different errors
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Write([]byte("success"))
	return
}
