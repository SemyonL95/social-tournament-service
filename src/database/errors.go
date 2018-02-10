package database

import "errors"

var ErrNotEnoughMoney = errors.New("not enough money")
var ErrUserAlreadyJoined = errors.New("user already joined in tournament")
var ErrNoPlayersInTournament = errors.New("no players in tournaments")