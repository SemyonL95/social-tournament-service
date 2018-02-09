package controllers

import (
	"net/http"
	"strconv"
	"database/sql"
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
	"regexp"

	"github.com/lib/pq"

	"github.com/SemyonL95/social-tournament-service/src/database"
	"github.com/SemyonL95/social-tournament-service/src/validators"
	"github.com/SemyonL95/social-tournament-service/src/models"
)

var BackersRegExp = regexp.MustCompile(`backerId=[A-Za-z0-9]*`)

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
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	parsedId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if parsedId <= 0 {
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
	rawBackersIds := BackersRegExp.FindAllString(r.URL.RawQuery, -1)
	tournamentId := r.FormValue("tournamentId")
	username := r.FormValue("playerId")

	var backersIds []string
	for _, str := range rawBackersIds {
		splitedStr := strings.Split(str, "=")
		validated := validators.ValidateUsername(splitedStr[1])
		if !validated {
			http.Error(w, "backerId have to be a string A-Za-z0-9 min: 1, max: 20 characters", http.StatusUnprocessableEntity)
			return
		}
		backersIds = append(backersIds, splitedStr[1])
	}

	if tournamentId == "" {
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	parsedTournamentId, err := strconv.Atoi(tournamentId)
	if err != nil {
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if parsedTournamentId <= 0 {
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	validated := validators.ValidateUsername(username)
	if !validated {
		http.Error(w, "playerId is required and have to be a string A-Za-z0-9 min: 1, max: 20 characters \n", http.StatusUnprocessableEntity)
		return
	}

	err = db.AssignToTournament(parsedTournamentId, username, backersIds)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			http.Error(w, err.Error(), http.StatusNotFound)
			break
		case database.ErrNotEnoughMoney:
			http.Error(w, err.Error(), http.StatusForbidden)
			break
		case database.ErrUserAlreadyJoined:
			http.Error(w, err.Error(), http.StatusForbidden)
			break
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Write([]byte("success"))
	return
}

func resultTournament(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	winners := models.Winners{}
	err = json.Unmarshal(b, &winners)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if winners.TournamentID == 0 {
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	for _, winner := range winners.Winners {
		validated := validators.ValidateUsername(winner.PlayerID)
		if !validated {
			http.Error(w, "backerId have to be a string A-Za-z0-9 min: 1, max: 20 characters", http.StatusUnprocessableEntity)
			return
		}

		validated = validators.ValidateFloatNotNegative(winner.Prize)
		if !validated {
			http.Error(w, "prize is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
			return
		}
	}

	err = db.FinishTournament(winners)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("success"))
	return
}

func reset(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	err := db.Truncate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("success"))
	return
}
