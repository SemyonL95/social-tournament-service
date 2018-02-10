package main_test

import (
	"gopkg.in/h2non/baloo.v3"
	"testing"
)

var test = baloo.New("http://app")

func TestCase(t *testing.T) {
	test.Get("/reset").
		Expect(t).
		Status(200).
		Done()

	fundUsers(t)
	announceTournament(t)
	joinTournament(t)
	balanceUsers(t)

}

func fundUsers(t *testing.T) {
	test.Get("/fund").
		AddQuery("playerId", "P1").
		AddQuery("points", "300").
		Expect(t).
		Status(200).
		Done()

	test.Get("/fund").
		AddQuery("playerId", "P2").
		AddQuery("points", "300").
		Expect(t).
		Status(200).
		Done()

	test.Get("/fund").
		AddQuery("playerId", "P3").
		AddQuery("points", "300").
		Expect(t).
		Status(200).
		Done()

	test.Get("/fund").
		AddQuery("playerId", "P4").
		AddQuery("points", "500").
		Expect(t).
		Status(200).
		Done()

	test.Get("/fund").
		AddQuery("playerId", "P5").
		AddQuery("points", "1000").
		Expect(t).
		Status(200).
		Done()
}

func announceTournament(t *testing.T) {
	test.Get("/announceTournament").
		AddQuery("tournamentId", "1").
		AddQuery("deposit", "1000").
		Expect(t).
		Status(200).
		Done()
}

func joinTournament(t *testing.T) {
	test.Get("/joinTournament").
		AddQuery("tournamentId", "1").
		AddQuery("playerId", "P5").
		Expect(t).
		Status(200).
		Done()

	test.Get("/joinTournament").
		AddQuery("tournamentId", "1").
		AddQuery("playerId", "P1").
		AddQuery("backerId", "P2").
		AddQuery("backerId", "P3").
		AddQuery("backerId", "P4").
		Expect(t).
		Status(200).
		Done()

}

func balanceUsers(t *testing.T) {
	test.Post("/resultTournament").
		JSON(`{
						"tournamentId": 1,
						"winners": [
							{"playerId": "P1", "prize": 2000}
							]
					}`).
		Expect(t).
		Status(200).
		Done()

	test.Get("/balance").
		AddQuery("playerId", "P1").
		Expect(t).
		Status(200).
		Status(200).
		JSON(`{
				"balance": "550.00",
				"playerId": "P1"
			}`).
		Done()

	test.Get("/balance").
		AddQuery("playerId", "P2").
		Expect(t).
		Status(200).
		Status(200).
		JSON(`{
				"balance": "550.00",
				"playerId": "P2"
			}`).
		Done()

	test.Get("/balance").
		AddQuery("playerId", "P3").
		Expect(t).
		Status(200).
		JSON(`{
				"balance": "550.00",
				"playerId": "P3"
			}`).
		Done()

	test.Get("/balance").
		AddQuery("playerId", "P4").
		Expect(t).
		Status(200).
		JSON(`{
				"balance": "750.00",
				"playerId": "P4"
			}`).
		Done()

	test.Get("/balance").
		AddQuery("playerId", "P5").
		Expect(t).
		Status(200).
		JSON(`{
				"balance": "0.00",
				"playerId": "P5"
			}`).
		Done()
}
