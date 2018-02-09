package database

const queryUpdateUsersCredits = "UPDATE users SET points = :points WHERE id = :id"
const querySelectUserForUpdate = "SELECT * FROM users WHERE id = $1 FOR UPDATE"
