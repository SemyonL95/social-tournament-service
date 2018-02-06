CREATE TABLE IF NOT EXISTS users (
  id       SERIAL PRIMARY KEY UNIQUE,
  username VARCHAR(255)   NOT NULL UNIQUE,
  credits  NUMERIC(15, 2) NOT NULL
);
CREATE TABLE IF NOT EXISTS tournaments (
  id        SERIAL PRIMARY KEY UNIQUE,
  winner_id INT NULL,
  deposit   NUMERIC(15, 2),
  CONSTRAINT foreign_winner_id FOREIGN KEY (winner_id) REFERENCES users (id)
);
CREATE TABLE IF NOT EXISTS players (
  id      SERIAL PRIMARY KEY UNIQUE,
  user_id INT NOT NULL,
  deposit NUMERIC(15, 2),
  CONSTRAINT foreign_user_id FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE TABLE IF NOT EXISTS bakers (
  id        SERIAL PRIMARY KEY UNIQUE,
  player_id INT NOT NULL,
  baker_id  INT NOT NULL,
  deposit   NUMERIC(15, 2),
  CONSTRAINT foreign_baker_id FOREIGN KEY (baker_id) REFERENCES users (id),
  CONSTRAINT foreign_player_id FOREIGN KEY (player_id) REFERENCES users (id)
);