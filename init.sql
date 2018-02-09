CREATE TABLE IF NOT EXISTS users (
  id     VARCHAR(255) PRIMARY KEY UNIQUE,
  points NUMERIC(15, 2) NOT NULL
);
CREATE TABLE IF NOT EXISTS tournaments (
  id       INT PRIMARY KEY UNIQUE,
  finished BOOLEAN DEFAULT FALSE NOT NULL,
  deposit  NUMERIC(15, 2)
);
CREATE TABLE IF NOT EXISTS players (
  id            SERIAL PRIMARY KEY UNIQUE,
  user_id       VARCHAR(255) NOT NULL,
  tournament_id INT          NOT NULL,
  CONSTRAINT foreign_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT foreign_tournament_id FOREIGN KEY (tournament_id) REFERENCES tournaments (id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS backers (
  id            SERIAL PRIMARY KEY UNIQUE,
  player_id     VARCHAR(255) NOT NULL,
  backer_id     VARCHAR(255) NOT NULL,
  tournament_id INT          NOT NULL,
  CONSTRAINT foreign_backer_id FOREIGN KEY (backer_id) REFERENCES users (id),
  CONSTRAINT foreign_player_id FOREIGN KEY (player_id) REFERENCES users (id),
  CONSTRAINT foreign_tournament_id FOREIGN KEY (tournament_id) REFERENCES tournaments (id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS results (
  id            SERIAL PRIMARY KEY UNIQUE,
  tournament_id INT          NOT NULL,
  winner_id     VARCHAR(255) NOT NULL,
  prize         NUMERIC(15, 2),
  CONSTRAINT foreign_winner_id FOREIGN KEY (winner_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT foreign_tournament_id FOREIGN KEY (tournament_id) REFERENCES tournaments (id) ON DELETE CASCADE
);