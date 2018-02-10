package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"github.com/SemyonL95/social-tournament-service/src/database"
	"github.com/SemyonL95/social-tournament-service/src/models"
	"github.com/SemyonL95/social-tournament-service/src/validators"
)

var BackersRegExp = regexp.MustCompile(`backerId=[A-Za-z0-9]*`)

func fund(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("playerId")
	points := r.FormValue("points")

	if ok := validators.ValidateUsername(username, "playerId", w); !ok {
		return
	}

	parsedPoints, err := strconv.ParseFloat(points, 64)
	if err != nil {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if ok := validators.ValidateFloatNotNegative(parsedPoints, "points", w); !ok {
		return
	}

	err = db.FundOrCreateUser(username, parsedPoints)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("user " + username + " funded"))
	return
}

func take(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if ok := validators.ValidateMethod("GET", r, w); !ok {
		return
	}

	username := r.FormValue("playerId")
	points := r.FormValue("points")

	if ok := validators.ValidateUsername(username, "playerId", w); !ok {
		return
	}

	parsedPoints, err := strconv.ParseFloat(points, 64)
	if err != nil {
		http.Error(w, "points is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if ok := validators.ValidateFloatNotNegative(parsedPoints, "points", w); !ok {
		return
	}

	err = db.TakePointsFromUser(username, parsedPoints)
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
	if ok := validators.ValidateMethod("GET", r, w); !ok {
		return
	}

	id := r.FormValue("tournamentId")
	deposit := r.FormValue("deposit")

	parsedId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "tournamentId is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if ok := validators.ValidateIntNotNegative(parsedId, "tournamentId", w); !ok {
		return
	}

	parsedDeposit, err := strconv.ParseFloat(deposit, 64)
	if err != nil {
		http.Error(w, "deposit is required and points have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if ok := validators.ValidateFloatNotNegative(parsedDeposit, "deposit", w); !ok {
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
	if ok := validators.ValidateMethod("GET", r, w); !ok {
		return
	}
	rawBackersIds := BackersRegExp.FindAllString(r.URL.RawQuery, -1)
	tournamentId := r.FormValue("tournamentId")
	username := r.FormValue("playerId")

	var backersIds []string
	for _, str := range rawBackersIds {
		splitedStr := strings.Split(str, "=")
		if ok := validators.ValidateUsername(splitedStr[1], "playerId", w); !ok {
			return
		}
		backersIds = append(backersIds, splitedStr[1])
	}

	parsedTournamentId, err := strconv.Atoi(tournamentId)
	if err != nil {
		http.Error(w, "id is required and id have to be numeric and not negative \n", http.StatusUnprocessableEntity)
		return
	}

	if ok := validators.ValidateIntNotNegative(parsedTournamentId, "tournamentId", w); !ok {
		return
	}

	if ok := validators.ValidateUsername(username, "playerId", w); !ok {
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
	if ok := validators.ValidateMethod("POST", r, w); !ok {
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

	if ok := validators.ValidateIntNotNegative(winners.TournamentID, "tournamentId", w); !ok {
		return
	}

	for _, winner := range winners.Winners {
		if ok := validators.ValidateUsername(winner.PlayerID, "playerId", w); !ok {
			return
		}

		if ok := validators.ValidateFloatNotNegative(winner.Prize, "prize", w); !ok {
			return
		}
	}

	err = db.FinishTournament(winners)
	if err != nil {
		switch err {
		case database.ErrNoPlayersInTournament:
			http.Error(w, err.Error(), http.StatusNotFound)
			break
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Write([]byte("success"))
	return
}

func reset(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if ok := validators.ValidateMethod("GET", r, w); !ok {
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

func balance(db *database.DB, w http.ResponseWriter, r *http.Request) {
	if ok := validators.ValidateMethod("GET", r, w); !ok {
		return
	}
	username := r.FormValue("playerId")
	if ok := validators.ValidateUsername(username, "playerId", w); !ok {
		return
	}
	user, err := db.GetUserBalance(username)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			http.Error(w, err.Error(), http.StatusNotFound)
			break
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	tmpUser := map[string]interface{}{
		"balance":  user.Points,
		"playerId": user.ID,
	}

	res, err := json.Marshal(&tmpUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
	return
}
