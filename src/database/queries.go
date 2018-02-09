package database

const queryUpdateUsersCredits = "UPDATE users SET points = :points WHERE username = :username"
const querySelectUserForUpdate = "SELECT * FROM users WHERE username = $1 FOR UPDATE"